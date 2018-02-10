# Social Tournament Service

### Task:

[Test task link](http://www.viktor.ee/backend-assessment.pdf)

### Description how to setup project:

1. Pull the project
2. Run command "docker-compose build"
3. Edit your /etc/host file and add app url, for example my is:
```
127.0.0.1       app
```
4. Run command "docker-compose up -d"

And now your project serving on http://app and also available via localhost

#### NOTICE!
in endpoint #4 (/resultTournament) added new param tournamentId, example:
```

{
    "tournamentId": 1
    "winners": [
        {"playerId": "P1", "prize": 2000}
    ]
}

```

## TESTS
I have e2e test which cover use case from test task, for run test just run command:

 go test ./src

### P.S.
 I added .env file into the repo, i know that it's not correct to do things like this, but it for simplification of project deploy