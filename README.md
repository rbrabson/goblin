# goblin

A work-in-progress to learn more about Discord bots and MongoDB, using the Go programming language.

## Running the goblin Bot

### Configuring the goblin Bot

The goblin bot relies on a set of environment variables to configure it.

#### goblin Bot

```bash
# Discord Bot Configuration
BOT_TOKEN="<bot_token>"
APP_ID="<bot_application_id>"

# You can use this variable to point at a development server, in which case any
# changes you have made will only appear on the development server.
# DISCORD_GUILD_ID="<server ID>"

# goblin DB configuration. This example shows you how to connect to MongoDB within a
# container, where the name of the deployed MongoDB container is `DISCORD_mongo`. If 
# running outside a container, replace `DISCORD_mongo` with the IP address or DNS name
# of your MongoDB instance. Prior to goblin being able to connect to the MongoDB
# instance, create the MongoDB database and add a database user for the database
# with Read/Write permissions, and use those values below.
GOBLIN_DEFAULT_THEME="clash"

# Heist DB configuration
MONGODB_SERVER="hostname"
MONGODB_USERID="username"
MONGODB_PASSWORD="password"
MONGODB_DATABASE="database"

# Heist DB URI
MONGODB_URI="mongodb+srv://$MONGODB_USERID:$MONGODB_PASSWORD@$MONGODB_SERVER/$MONGODB_DATABASE?retryWrites=true&w=majority"

# For production environmenbts, don't set DISCORD_GUILD_ID, but it can be useful
# when configurinig the guild for sting or debugging. This will only register
# the new commands with the specific server that has this ID assigned.
# Note that there is a limit to how many times per day you can update the
# commands, so if you find that Discord is not responding to your bot's command
# registrations, you have have hit this limit.
DISCORD_GUILD_ID="<server-id>"
```

#### Configuring MongoDB for the goblin Bot

The MongoDB database needs to have a user configured who can read and write from the goblin database. Using the
`mongosh` command to connect to your MongoDB instance, including any username/password credentials that may be
required, you can add a user by using the following command. For example, you may need to specify a command
such as this if the MongoDB instance is running locally:

```bash
 'localhost:27017' -u <root_username> -p <root_password>
```

Or, if you are running mongodb remotely:

```bash
mongosh -host <ip_address>:<port> -u <root_username> -p <root_password>
```

Once mongosh has started, enter the following command to create a user who can read and write to the specified
database.

```bash
use admin
db.createUser(
  {
    user: "<DISCORD_db_userid>",
    pwd: "<DISCORD_db_pwd>",
    roles: [ { role: "readWrite", db: "<DISCORD_db_name>" } ]
  }
)
```

Note that the actual MongoDB database will be created when the first collection or document is written to the database.

### Run as a Standalone Application

When developing, you can use

```bash
go run cmd/goblin/main.go
```

to compile and run the goblin bot. Once it is stable, you can use the `make` command to generate a binary that you can
execute.

### Run as a Docker Image

#### Build Container

You should edit your `.env` file to set `DISCORD_STORE="mongodb"` when deploying in this manner. You should manually start
a MongoDB instance that your deployed image may use for access.

``` bash
docker-buildx build -t goblin:1.0.0 .
docker push goblin:1.0.0
```

#### Start Container

```bash
docker run --envfile ./.env --name <container-name> goblin:1.0.0
```

### Run using `docker compose`

The following command will both build the container, as well as deploy with both the goblin bot as well as MongoDB. You should edit your
`.env` file to set `DISCORD_STORE="mongodb"` when deploying in this manner.

```bash
docker compose up --build
```

### Run in Pterodactyl

Pterodactyl is a game server management pane that runs all game servers in isolated Docker containers.

#### Define the Egg

##### Specify the configuration variables

With Pterodactyl, you need to create an `egg` that defines the goblin bot. This `egg` is then placed in
a `nest`. For example, you might have a `Discord` nest, and then create the goblin `egg` within that nest.

For goblin, the first step is to create an egg for the `generic golang application`. This egg requires
configuration in order to be able to run.

In the Pterodactyl interface, navigate to the `Nest` section and select the `egg` you created. You should
include the options defined above for the bot.

- BOT_TOKEN. This is a required string value.

- APP_ID. This is a required string value.

- GOBLIN_DEFAULT_THEME. This is a required string value. It should default to `clash`.

- MONGODB_SERVER. The server to which to connect.

- MONGODB_USERID. The user ID for the bot to access the server.

- MONGODB_PASSWORD. The password for the ot to access the server.

- MONGODB_DATABASE. The databse in which the bot's data is stored.

- MONGODB_URI. A URI built using the other MONGO_xxxx values set above. It is set to `"mongodb+srv://$MONGODB_USERID:$MONGODB_PASSWORD@$MONGODB_SERVER/$MONGODB_DATABASE?retryWrites=true&w=majority"`

##### Configure the startup script

Under the egg, configure the startup script to look like the following.

```bash
#!/bin/bash
# golang generic package

if [ ! -d /mnt/server/ ]; then
    mkdir -p /mnt/server/
fi

# Download and install a more recent version of go. The one that is part
# of the golang generic package is too old.
wget https://go.dev/dl/go1.20.6.linux-amd64.tar.gz
rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.6.linux-amd64.tar.gz
rm -f go1.20.6.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Clone the code from github so that it can be built
git clone https://github.com/rbrabson/goblin.git

# Move into the bot directory
cd goblin

# Download the dependencies, both direct and indirect, required to build
# the package
go mod download

# Use a local tmp directory. The global one for this server was too small.
mkdir ~/tmp
export TMPDIR=~/tmp

# Build the linux binary image
make build-linux

# Copy the image to the correct location
cp -f bin/linux/amd64/goblin /mnt/server/
```

#### Install the server

Under the server, configure the specific values for the bot. Once done, you can re-install
the bot, and then reinstall the bot. Once the bot is re-installed, you can start the bot.

If the bot is already running, stop it before reinstalling.
