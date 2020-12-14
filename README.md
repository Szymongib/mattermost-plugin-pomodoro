# Mattermost Pomodoro Plugin

Mattermost plugin that allows you to run Pomodoro sessions from Mattermost.

This is proof-of-concept implementation.

## Features

To start the Pomodoro timer use slash commands:
```
/pomodoro start 30m
```

The plugin will automatically set your status to `do not disturb` and Pomodoro bot will notify you when the time has passed.

You can list your past sessions by typing:
```
/pomodoro list
```
The bot will deliver them in the DM.
