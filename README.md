# hipcat

hipcat is a command line tool that posts messages to [Hipchat] This tool was forked from [slackcat][slackcat] to support Hipchat.

    $ echo "hello" | hipcat

## Installation

If you have a working Go installation run `go get github.com/jburnham/hipcat`.

## Configuration

You need to create an [API Token][new-api-token].

You can then configure hipcat through a config file and/or environment variables.

### JSON File

```json
{
    "api_token":"RK4U0ZmdsTkc4ap8tKY9qDe8Ps1lT38cBh2ohGaZ",
    "hipchat_url":"https://api.hipchat.com",
    "room":"MyRoom"
}
```

In `/etc/hipcat.conf`, `~/.hipcat.conf` or `./hipcat.conf`

See `hipcat-example.conf` for
a full example.


### Environment Variable

```bash
$ export HIPCAT_API_TOKEN=RK4U0ZmdsTkc4ap8tKY9qDe8Ps1lT38cBh2ohGaZ
$ export HIPCAT_URL="https://api.hipchat.com"
$ export HIPCAT_ROOM="MyRoom"
```

Will override file config.

## Usage

hipcat will take each line from stdin and post it as a message to Slack:

    tail -F logfile | hipcat

Be aware that if a file outputs blank lines, this will result in a 400 error from Hipchat. You can remedy this using
grep to filter out blank lines:

    tail -F logfile | grep --line-buffered -v '^\s*$' | hipcat

If you'd prefer to provide a message on the command line, you can:

    sleep 300; hipcat "done"

### Room
Default: None

    echo "sudo make me a sandwich" | hipcat --room test

[Hipchat]: https://www.hipchat.com
[slackcat]: https://github.com/skattyadz/slackcat
[new-api-token]: https://www.hipchat.com/docs/apiv2/auth
