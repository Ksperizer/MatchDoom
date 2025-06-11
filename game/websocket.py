import asyncio 
import websockets
import json
import logging

logging.basicConfig(level=logging.INFO)

connected_players = []

async def handle_player(websocket, path):
    try: 
        logging.info("Nouvelle connexion")

        message = await websocket.recv()
        data = json.loads(message)

        if data.get("type") != "join" or not data.get("pseudo"):
            await websocket.send(json.dumps({"type": "error", "message": "Données invalides"}))
            return
        
        pseudo = data["pseudo"]
        logging.info(f"{pseudo} a rejoint la file d'attente")
        
        connected_players.append((pseudo, websocket))

        # notify player position in queue
        await websocket.send(json.dumps({
            "type":  "queue",
            "message": "Ajouter  à la file d'attente",
            "position": len(connected_players)
        }))

        # if 2 players are connected, start the game
        if len(connected_players) == 2:
            p1, ws1 = connected_players.pop(0)
            p2, ws2 = connected_players.pop(0)

            # notify players that the game is starting
            await ws1.send(json.dumps({
                "type": "game_start",
                "opponent": p2
            }))
            await ws2.send(json.dumps({
                "type": "game_start",
                "opponent": p1
            }))

            logging.info(f"Le match commence {p1} vs {p2}")

    except websockets.exceptions.ConnectionClosed:
        logging.warning("Connexion fermée")
    except Exception as e:
        logging.error(f"Erreur serveur : {e}")

# start the server
async def main():
    async with websockets.serve(handle_player, "0.0.0.0", 8081):
        logging.info("Serveur WebSocket démarré sur ws://localhost:8081")
        await asyncio.Future() # run forever

if __name__ == "__main__":
    asyncio.run(main())
            