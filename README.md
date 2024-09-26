# dgame

A work-in-progress to learn more about Discord bots and MongoDB, using the Go programming language.

## Running the dgame Bot

### Configuring the dgame Bot

The dgame bot relies on a set of environment variables to configure it.

#### dgame Bot

```bash
# Discord Bot Configuration
BOT_TOKEN="<bot_token>"
APP_ID="<bot_application_id>"

# dgame State Storage Method. Choose one or the other. If you are deploying to a
# container, then you should be using "mongodb", otherwise the data will not be
# persisted.
# DISCORD_STORE="file"
DISCORD_STORE="mongodb"

# dgame File Store Configuration. If running within a container, you must use the
# values below. If running as a stand-alone application, you can change them to your
# preferred location.
DISCORD_FILE_STORE_DIR="./store/"

# You can use this variable to point at a development server, in which case any
# changes you have made will only appear on the development server.
# DISCORD_GUILD_ID="<server ID>"

# dgame DB configuration. This example shows you how to connect to MongoDB within a
# container, where the name of the deployed MongoDB container is `dgame_mongo`. If 
# running outside a container, replace `dgame_mongo` with the IP address or DNS name
# of your MongoDB instance. Prior to dgame being able to connect to the MongoDB
# instance, create the MongoDB database and add a database user for the database
# with Read/Write permissions, and use those values below.
MONGODB_URI="mongodb+srv://<username>:<password>@<uri>/<database>?retryWrites=true&w=majority"

# For production environmenbts, don't set DISCORD_GUILD_ID, but it can be useful
# when configurinig the guild for sting or debugging. This will only register
# the new commands with the specific server that has this ID assigned.
# Note that there is a limit to how many times per day you can update the
# commands, so if you find that Discord is not responding to your bot's command
# registrations, you have have hit this limit.
DISCORD_GUILD_ID="<server-id>"
```

#### Configuring MongoDB for the dgame Bot

The MongoDB database needs to have a user configured who can read and write from the dgame database. Using the
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
    user: "<dgame_db_userid>",
    pwd: "<dgame_db_pwd>",
    roles: [ { role: "readWrite", db: "<dgame_db_name>" } ]
  }
)
```

Note that the actual MongoDB database will be created when the first collection or document is written to the database.

### Run as a Standalone Application

When developing, you can use

```bash
go run cmd/dgame/main.go
```

to compile and run the dgame bot. Once it is stable, you can use the `make` command to generate a binary that you can
execute.

### Run as a Docker Image

#### Build Container

You should edit your `.env` file to set `DISCORD_STORE="mongodb"` when deploying in this manner. You should manually start
a MongoDB instance that your deployed image may use for access.

``` bash
docker-buildx build -t dgame:1.0.0 .
docker push dgame:1.0.0
```

#### Start Container

```bash
docker run --envfile ./.env --name <container-name> dgame:1.0.0
```

### Run using `docker compose`

The following command will both build the container, as well as deploy with both the dgame bot as well as MongoDB. You should edit your
`.env` file to set `DISCORD_STORE="mongodb"` when deploying in this manner.

```bash
docker compose up --build
```

### Run in Pterodactyl

Pterodactyl is a game server management pane that runs all game servers in isolated Docker containers.

#### Define the Egg

##### Specify the configuration variables

With Pterodactyl, you need to create an `egg` that defines the dgame bot. This `egg` is then placed in
a `nest`. For example, you might have a `Discord` nest, and then create the dgame `egg` within that nest.

For dgame, the first step is to create an egg for the `generic golang application`. This egg requires
configuration in order to be able to run.

In the Pterodactyl interface, navigate to the `Nest` section and select the `egg` you created. You should
include the options defined above for the bot.

- BOT_TOKEN. This is a required string value.

- APP_ID. This is a required string value.

- DISCORD_STORE. This is is a required string value. It should default to `mongodb`, but can also be `file`

- DISCORD_FILE_STORE_DIR. This is an optional string value, but required if DISCORD_STORE is set to `file`. It should default to `./store/`.

- DISCORD_DEFAULT_THEME. This is a required string value. It should default to `clash`.

- MONGODB_URI. This is an optional string value, but required if DISCORD_STORE is set to `mongo`.

- MONGODB_USERID. This is an optional string value, but required if DISCORD_STORE is set to `mongo`.

- MONGODB_PASSWORD. This is an optional string value, but required if DISCORD_STORE is set to `mongo`.

- MONGODB_DATABASE. This is an optional string value, but required if DISCORD_STORE is set to `mongo`.

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
git clone https://github.com/rbrabson/dgame.git

# Move into the bot directory
cd dgame

# Download the dependencies, both direct and indirect, required to build
# the package
go mod download

# Use a local tmp directory. The global one for this server was too small.
mkdir ~/tmp
export TMPDIR=~/tmp

# Build the linux binary image
make build-linux

# Copy the image to the correct location
cp -f bin/linux/amd64/dgame /mnt/server/
```

#### Install the server

Under the server, configure the specific values for the bot. Once done, you can re-install
the bot, and then reinstall the bot. Once the bot is re-installed, you can start the bot.

If the bot is already running, stop it before reinstalling.
