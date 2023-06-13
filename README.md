# tiktok_tech_immersion_2023

![Tests](https://github.com/jianminglok/assignment_demo_2023/actions/workflows/test.yml/badge.svg)

This is a completed version of the backend assignment of 2023 TikTok Tech Immersion. It composes of a HTTP server and a RPC server connected to a MySQL database. Users are able to send and pull message(s) via HTTP requests through the API endpoints listed [below](#instructions).

## Installation

Requirement:

- [Golang 1.18+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

You may also wish to install K6 if you want to run the tests under ```/tests```.

## Getting started

```bash
docker-compose up -d
```
> It is recommended to clone the project directly under a Linux/Unix system even though Windows supports docker-compose through WSL2/Hyper-V, due to potential issues with the carriage-return character when cloning on Windows and subsequently running docker-compose via WSL2.

## Rebuilding after making changes

```bash
docker-compose up -d --build
```

## Instructions
- All requests sent to the HTTP server should have a valid JSON body

- API Endpoints:
    - Sending messages ```/api/send```
        | Parameter      | Description | Format | Type | Required |
        | ----------- | ----------- | ----------- | ----------- | ----------- | 
        | chat      | Chat to send message to       | personA:personB, name of person should not contain ```:``` | string | Yes |
        | text   | Message to send        | N/A | string | Yes |
        | sender | Name of sender, should not contain ```:``` | personA | string | Yes |

        Example:

        ```
        {
            "chat":"personA:personB",
            "text": "text to be sent", 
            "sender": "personA"
        }
        ```

    - Pulling messages 
        | Parameter      | Description | Format | Type | Default | Required |
        | ----------- | ----------- | ----------- | ----------- | ----------- | ----------- | 
        | chat      | Chat to pull messages from      | personA:personB, name of person should not contain ```:``` | string | N/A | Yes |
        | cursor   | Minimum sendtime of messages to pull        | N/A | int64 | 0 | Optional |
        | limit | Maximum number of messages to pull | N/A | int32 | 10 | Optional |
        | reverse | Messages will be sorted in descending accorder by sendtime if set to true | N/A | boolean | false | Optional |

        Example:

        ```
        {
            "chat":"personA:personB",
            "cursor": 1686593248656448855,
            "limit": 5, 
            "reverse": true
        }
        ```

## Tech Stack
- Golang
- Kitex
- Hertz
- Docker
- MySQL
- K6