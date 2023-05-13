# `logspark` ✨

Run shell commands based on regex pattern matches in one or more continuously read files.

## Running `logspark`

```sh
  logspark <path to configuration file>
```

## Configuration

`logspark` is configured using a TOML file as follows:

```toml
logging = "minimal"
# logging can be set to "none", "minimal", or "verbose"
# - none: disables logging
# - minimal: <timestamp> | <file path> | <RegEx pattern name> | <matched line>
# - verbose: <timestamp> | <file path> | <RegEx pattern name> | <RegEx pattern> | <command that ran> | <alert message> | <matched line>

log_file = "none"
# log_file can be set to "none", "stdout", or any valid path
# - none: no log file
# - stdout: print to stdout (console)
# - /path/to/log/file: the path to the desired log file (is created if it doesn't exist)

files = [ "/path/to/file", "/var/log/dmesg" ] # must be an array
# These are the tailed files

[[regex]]
name = "Name of RegEx pattern"
regex = '''RegEx pattern''' #must be a literal string (no TOML escaping)
command = '''command to run when the pattern is matched'''
alert = '''same as above, but used for alerting'''

[[regex]]
name = "reboot"
regex = '''reboot'''
command = '''sleep 60; reboot'''
alert = '''wall 'Rebooting in 60 seconds!''''
```

You can watch as many files and search for as many regex patterns as you like.
