# dcdl

This is a Discord bot for downloading attachments from channel messages.

## Setup

First, set up a Discord bot application (see [here](https://discordjs.guide/preparations/setting-up-a-bot-application.html#creating-your-bot) how to do so).

Now, you can go to `OAuth2` in the Discord bot application page and go to the sub route `URL Generator`. There, select the scopes `bot` and `applications.commands`. Now, you can copy the `Generated Url` and pate it into your browser. You might now log in to Discord. After that, select the server you want to have the bot on and invite it.

### Using Docker

You can use the provided [`docker-compose.yml`](docker-compose.yml) to set up the bot. In the `environment` configuration, you must set the bot token obtained from the Discord bot application.

You can of course also run the image using the Docker CLI as well.
```
$ docker run -d \
    --publish "80:80" \
    --name dcdl \
    --env "DCDL_DISCORD_TOKEN=OTEwN..." \
    --env "DCDL_WEBSERVER_PUBLICADDRESS=https://example.com" \
    ghcr.io/zekrotja/dcdl:latest
```

### Using Precompiled Binaries

You can also download precompiled binaries from the [actions pipeline](https://github.com/zekroTJA/dcdl/actions/workflows/artifacts.yml).

Then, simply create a configuration file.

> ./config.yml
```yaml
discord:
  token: OTEwN...

webserver:
  publicaddress: https://example.com

storage:
  location: ./data
```

Following, cerate the data directory.
```
$ mkdir data
```

After that, start the bot with the following command line.
```
$ ./dcdl -c config.yml
```

## Usage

Now, go into a text channel and type `/collect`. Then, the command should pop up.

![image](https://user-images.githubusercontent.com/16734205/142413107-59374f21-36d3-4512-abaf-b12a95435e6f.png)

Here, you can now specify following optional arguments.

- `channel`: The target channel. This is set to the current channel if not specified.

- `limit`: The maximum amount of messages to be fetched. This defaults to `0` (equals all messages in the channel) or the globally set limit.

- `include-metadata`: Whether or not to include a `metadata.json` in the colelction package containing information about each message containing attachments. This defaults to `true`.

- `include-files`: Whether or not download and include the attachment files. You can set this to `false` if you want to download the files from your system using the `metadata.json` file. This defaults to `true`.

## Download via `metadata.json`

You can also download the attachments on your system using the `metadata.json` file from the archive using the following command.

> Therefore, you need `curl` and `jq` installed. If you are on windows, use WSL. 😉
```
$ mkdir files && cat metadata.json | jq -r '.[].attachments[] | [ .archive_filename, .url ] | join(" ")' | xargs -l bash -c 'curl -Lo "files/$0" "$1"'
```

You can even do stuff like filter by author ID, for example, using the following command.
```
$ cat metadata.json | jq -r '.[] | select( .author_id == "221905671296253953" )
```

... and then combine it with the download command to just download attachments sent by that specific user.

```
$ mkdir files && cat metadata.json | jq -r '.[] | select( .author_id == "221905671296253953" ) | .attachments[] | [ .archive_filename, .url ] | join(" ")' | xargs -l bash -c 'curl -Lo "files/$0" "$1"'
```

*Man, `jq` is really one of the most useful CLI tool ever created, inst it?* 😄