# corset

Squeeeeeze JSON to fit SCP limits.

Corset can: 

1. Remove whitepace from a JSON File (SCP, RCP, etc)
2. Unpack and pack policies efficiently across multiple files

... all to bring them under the magical 5120 character limit. 

## Installation
```bash
go install github.com/jakebark/tag-nag@latest
```
You may need to set [GOPATH](https://go.dev/wiki/SettingGOPATH).

## Commands

Remove the whitespace from a JSON file or files (in a directory). 

```bash
corset scp.json 
corset ./directory # run against a directory
```

Optional flags
```bash
-w # dont remove the whitespace
```

## Related Resources

- [AWS Organizations service quotas](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_reference_limits.html)
