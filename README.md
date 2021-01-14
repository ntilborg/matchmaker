# MatchMaker
Matchmaker example to incorporate into the [Agones](https://agones.dev/site/) kubernetes environment. The matchmaker is based on [kuy](https://github.com/faruqisan/kuy).

The matchmaker example incorporates the Agones kubernetes API to find a server and can be used in any environments by only adapting the `config.json` file.

# MatchMaker Agones example

Generate new UUIDs with `/register`:
````
# curl http://localhost:8001/register
3465284411
````

Join a match as a player (player id 1 and 2) with `/join`. Returns the match ID and some parameters. No match

````
# curl http://localhost:8001/join?id=1
{"MatchmakingID":1409079322,"IsFull":false,"CurrentPlayers":1,"MaxPlayers":5,"ServerPort":0,"ServerHost":""}
# curl http://localhost:8001/join?id=2
{"MatchmakingID":1409079322,"IsFull":false,"CurrentPlayers":2,"MaxPlayers":5,"ServerPort":0,"ServerHost":""}

````

For testing purposes you can automate multiple players to join:
````
# curl -s "http://localhost:8001/join?id=[1-10]"
````

Find match with `/match`:
````
# curl http://localhost:8001/match?id=1409079322
{"MatchmakingID":1409079322,"IsFull":true,"CurrentPlayers":5,"MaxPlayers":5,"ServerPort":7080,"ServerHost":"127.0.0.1"}
````

Please see the [agones/example](https://github.com/ntilborg/matchmaker/tree/master/example/agones) directory for more info


## License
[MIT](https://github.com/ntilborg/matchmaker/blob/master/LICENSE)