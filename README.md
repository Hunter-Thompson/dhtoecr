# dh2ecr 

## Deps

1. docker cli

## Install 

```
go install github.com/Hunter-Thompson/dh2ecr@latest
```

## Usage 
```
$ dh2ecr -r ap-south-1 -c example.yaml -a 123 -d true
```
```
Copies over images from dockerhub to ecr

Usage:
  dh2ecr [flags]

Flags:
  -a, --aws-account-id string   aws account id
  -r, --aws-region string       aws region
  -c, --config string           config file
  -d, --dry-run                 dry run
  -h, --help                    help for dh2ecr
```

## [Config](./example.yaml)
```yaml
registryMap:
  alpine:
    - alpine:latest
    - alpine:3.13
  nginx:
    - nginx:latest
```

## Logs 
```
$ dh2ecr -r ap-south-1 -c example.yaml -a 123 -d true
2022/04/06 01:43:16 reading config example.yaml
2022/04/06 01:43:16 config: {map[alpine:[alpine:latest alpine:3.13] nginx:[nginx:latest]]}
2022/04/06 01:43:16 pulling image alpine:latest from docker hub
2022/04/06 01:43:16 tagging image alpine:latest with tag 123.dkr.ecr.ap-south-1.amazonaws.com/alpine:latest
2022/04/06 01:43:16 pushing image 123.dkr.ecr.ap-south-1.amazonaws.com/alpine:latest to ecr
2022/04/06 01:43:16 pulling image alpine:3.13 from docker hub
2022/04/06 01:43:16 tagging image alpine:3.13 with tag 123.dkr.ecr.ap-south-1.amazonaws.com/alpine:3.13
2022/04/06 01:43:16 pushing image 123.dkr.ecr.ap-south-1.amazonaws.com/alpine:3.13 to ecr
2022/04/06 01:43:16 pulling image nginx:latest from docker hub
2022/04/06 01:43:16 tagging image nginx:latest with tag 123.dkr.ecr.ap-south-1.amazonaws.com/nginx:latest
2022/04/06 01:43:16 pushing image 123.dkr.ecr.ap-south-1.amazonaws.com/nginx:latest to ecr
```
