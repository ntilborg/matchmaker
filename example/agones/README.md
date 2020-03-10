# MatchMaker Agones example

Generate new IDs:
- http://localhost:8001/register

Join a match as a player (player id 1 and 2)
- http://localhost:8001/join?id=1
- http://localhost:8001/join?id=2

Automate multiple players to join
- curl -s "http://localhost:8001/join?id=[1-10]"

Find match
- http://localhost:8001/match?id=1000
