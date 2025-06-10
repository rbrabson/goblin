# Goblin

A Discord bot built with Go that integrates with MongoDB, providing various game and utility features for Discord servers.

## Changelog

See the [changelog](CHANGELOG.md) for details.

## Running the Goblin Bot

### Configuring the Goblin Bot

The Goblin bot relies on a set of environment variables to configure it.

#### Goblin Bot

```bash
# Discord Bot Configuration
DISCORD_BOT_TOKEN="<bot_token>"
DISCORD_APP_ID="<bot_application_id>"

# You can use this variable to point at a development server, in which case any
# changes you have made will only appear on the development server.
# DISCORD_GUILD_ID="<server ID>"

# Goblin DB configuration. This example shows you how to connect to MongoDB within a
# container, where the name of the deployed MongoDB container is `DISCORD_mongo`. If 
# running outside a container, replace `DISCORD_mongo` with the IP address or DNS name
# of your MongoDB instance. Prior to Goblin being able to connect to the MongoDB
# instance, create the MongoDB database and add a database user for the database
# with Read/Write permissions, and use those values below.
DISCORD_DEFAULT_THEME="clash"

# Heist DB configuration
MONGODB_USERID="<userid>"
MONGODB_PASSWORD="<password>"
MONGODB_SERVER="<server_host_name>"
MONGODB_DATABASE="database"

# Heist DB URI
MONGODB_URI="mongodb+srv://$MONGODB_USERID:$MONGODB_PASSWORD@$MONGODB_SERVER/$MONGODB_DATABASE?retryWrites=true&w=majority"

# For production environments, don't set DISCORD_GUILD_ID, but it can be useful
# when configuring the guild for testing or debugging. This will only register
# the new commands with the specific server that has this ID assigned.
# Note that there is a limit to how many times per day you can update the
# commands, so if you find that Discord is not responding to your bot's command
# registrations, you may have hit this limit.
DISCORD_GUILD_ID="<server-id>"

# Logging level for the bot. Options are "debug", "info", "warn", "error", "fatal"
# Default is "info"
LOG_LEVEL="info"
```

#### Configuring MongoDB for the Goblin Bot

The MongoDB database needs to have a user configured who can read and write from the configured database. Using the
`mongosh` command to connect to your MongoDB instance, including any username/password credentials that may be
required, you can add a user by using the following command. For example, you may need to specify a command
such as this if the MongoDB instance is running locally:

```bash
mongosh 'localhost:27017' -u <root_username> -p <root_password>
```

Or, if you are running MongoDB remotely:

```bash
mongosh --host <ip_address>:<port> -u <root_username> -p <root_password>
```

Once mongosh has started, enter the following command to create a user who can read and write to the specified
database.

```bash
use admin
db.createUser(
  {
    user: "<MONGODB_USERID>",
    pwd: "<MONGODB_PASSWORD>",
    roles: [ { role: "readWrite", db: "<MONGODB_DATABASE>" } ]
  }
)
```

Note that the actual MongoDB database will be created when the first collection or document is written to the database.

### Run as a Standalone Application

When developing, you can use:

```bash
go run cmd/goblin/main.go
```

to run the Goblin bot. Once it is stable, you can use the `make` command to generate a binary that you can
execute.

### Run as a Docker Image

#### Build Container

```bash
docker buildx build -t Goblin:1.0.0 .
docker push Goblin:1.0.0
```

#### Start Container

```bash
docker run --env-file ./.env --name <container-name> Goblin:1.0.0
```

### Run using `docker compose`

```bash
docker compose up --build
```

### Run in Pterodactyl

Pterodactyl is a game server management panel that runs all game servers in isolated Docker containers.

#### Define the Egg

##### Specify the configuration variables

With Pterodactyl, you need to create an `egg` that defines the Goblin bot. This `egg` is then placed in
a `nest`. For example, you might have a `Discord` nest, and then create the Goblin `egg` within that nest.

For Goblin, the first step is to create an egg for the `generic golang application`. This egg requires
configuration in order to be able to run.

In the Pterodactyl interface, navigate to the `Nest` section and select the `egg` you created. You should
include the options defined above for the bot.

- BOT_TOKEN. This is a required string value.

- APP_ID. This is a required string value.

- MONGODB_SERVER. The server to which to connect.

- MONGODB_USERID. The user ID for the bot to access the server.

- MONGODB_PASSWORD. The password for the bot to access the server.

- MONGODB_DATABASE. The database in which the bot's data is stored.

- MONGODB_URI. A URI built using the other MONGODB_xxxx values set above. It is set to `"mongodb+srv://$MONGODB_USERID:$MONGODB_PASSWORD@$MONGODB_SERVER/$MONGODB_DATABASE?retryWrites=true&w=majority"`

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

# Clone the code from GitHub so that it can be built
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
cp -f bin/linux/amd64/Goblin /mnt/server/
```

#### Install the server

Under the server, configure the specific values for the bot. Once done, you can reinstall the bot. After the bot is reinstalled, you can start it.

If the bot is already running, stop it before reinstalling.

## Configuring Discord

### Setting administrator permissions

In order to use administrative commands with the Goblin bot, users must have an administrative role assigned to them. When first installed, the Goblin bot assigns the following roles for admin users: `Admin`, `Admins`, `Administrator`, `Mod`, `Mods`, and `Moderator`. You can add additional roles or remove existing ones using the `/server role add` command, or remove an existing role using the `/server role remove` command. At any time, you can see the list of existing administrative roles using the `/server role list` command.

### Setting up Goblin permissions in the game channels

The Goblin bot requires the ability to mute and unmute the channel in which the `heist` game is run. To accomplish this, add the `Goblin` bot under the `ROLES/MEMBERS` section in the channel settings, and then add the following permissions to the bot: `View Channel`, `Manage Permissions`, and `Send Messages`. Without these settings, the `heist` bot will either be unable to mute the channel during a heist, or may be unable to send messages about the progress of a heist at all.

### Segregating administration and member commands to separate channels

There are two sets of commands used by the `Goblin` bot: administrative commands and member commands. All administrative commands start with `/admin-`, while the help command is `/adminhelp`. This allows a user to easily configure the integration settings for the Goblin bot to only allow the use of administrative commands in a channel accessible only to administrators. Failing to do so won't allow general users to use administrative commands, as this is protected via the administration roles. However, the administrative commands will show up in the list of available commands to the users, which may not be desirable.

### Setting the leaderboard channel

The leaderboard is posted once a month, at the start of the month using UTC time. To have the leaderboard posted, you must first set a channel to which the leaderboard is sent. This is accomplished using the `/lb-admin channel` command. Not doing so will result in no leaderboard being sent to your Discord server.
