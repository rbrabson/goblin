# Changelog

## 1.1.0-alpha1

Add the `race` game to the server. Races allow members to enter races, where they are assigned virtual racers to participate in the race. Racers who come in first, second or third earn prizes, with the prize amount increasing with the larger number of racers entering the race.

Members can also bet on the outcome of the race. If the racer on whom the racer bets wins the race, then their bet pays out winnings. The amount of winnings earned by picking the race winner increases with the number of members who enter the race.

## 1.0.0-alpha1

Initial Globlin bot, which is a rewrite of the older Heists bot.

- Heist game: participate in fictitious `heists` with other members of the server, attempting to steal credits from target vaults. Participaitng in a heist costs a non-refundable number of credits, and any ill-gotten gains are deposited into the player's bank accounts.

- **Leaderboard**

  - *Current*: displays members with the top 10 highest current bank account balance.

  - *Monthly*: displays members with the top 10 highest monthly bank account balance. The balance is reset at the start of each month.

  - *Lifetime*: displays members with the top 10 highest lifetime bank account balance. The balance is maintained even if withdrawls are made for a future `shop` (the shop does not currently exist)

- *Payday*: depsoit an amount of credits into the member's bank account. This can only be done once every 23 hours.

- *Bank*: the ability to display information about the member's bank account.

In addition, there are various administrative commands that are available to the game admninistrator.

- **Bank Admin**

  - **Account**: sets the bank account balance for a member of the server.

  - **Balance**: sets the default balance for the server. This is the amount of credits deposited into a user's bank account when it is first created. The intent is to allow a new member to be able to immediately play games using the bot when they first interact with it.

  - **Name**: Sets the name of the bank. This is only used when sending messages to the server.

  - **Currency**: Sets the name of the currency used on the bank. This is only used when sending messages to the server.

  - **Info**: Displays basic information about the banking system on the server.

- **Heist Admin**

  - **Configure**: Configure some basic aspects of the heist game.

  - **Theme**: Select the theme used by the heist game. The theme must already exist in order to be selected.

  - **Reset**: Reset a hung heist.

- **Leaderboard Admin**

  - **Channel**: Sets the channel to which the monthly leaderboard is sent. This must be done in order for a leaderboard to be sent monthly to the server.

  - **Info**: Displays basic information about the leaderboard configuration for the server.

- **Guild Admin**

  - **Role**: Add or remove an administrative role used to manage the bot. More than one role may be assigned. By default, the bot comes with various values that match commonly used *Administrator* and *Moderator* roles. It is recommended that a bot admin role is assigned to manage the bot, but it is not strictly required.
  
    Admins can also list the current set of administrative roles assigned to the bot.
