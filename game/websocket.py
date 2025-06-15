import asyncio 
import websockets
import json
import logging

logging.basicConfig(level=logging.INFO)

clients = set()
players_queue = []
game = {}

async def handle_player(websocket):
    logging.info("Nouvelle connexion webSocket {websocket.path}")
    clients.add(websocket)
    try:
        async for message in websocket: 
            logging.info(f"Message reçu: {message}")
            await websocket.send("echo: " + message)  # Echo the message back to the client
            try:
                data = json.loads(message)
                response = await handle_message(data, websocket)
                if response: 
                    await websocket.send(json.dumps(response))
            except json.JSONDecodeError:
                await websocket.send(json.dumps({"type": "error", "message": "Invalid JSON"}))
    except websockets.ConnectionClosed as e:
        logging.info(f"Connexion fermée: {e}")
    finally:
        clients.remove(websocket)

def remove_from_queue(ws):
    global players_queue
    players_queue = [p for p in players_queue if p["ws"] != ws]
    
async def handle_message(data, websocket):
    msg_type = data.get("type")

    if msg_type == "ping":
        await websocket.send(json.dumps({"type": "pong"}))

    elif msg_type == "join":
        pseudo = data.get("pseudo")
        logging.info(f"Pseudo {pseudo} a rejoint la file d'attente")

        # check if pseudo is provided
        if any(p["pseudo"] == pseudo for p in players_queue):
            await websocket.send(json.dumps({"type": "error", "message": "Pseudo déjà utilisé"}))
            return

        players_queue.append({"pseudo": pseudo, "ws": websocket})
        await websocket.send(json.dumps({
            "type": "queue",
            "message": "En attente d'un adversaire",
            "position": len(players_queue)
        }))

        if len(players_queue) >= 2:
            await start_game()

    else:
        await websocket.send(json.dumps({"type": "error", "message": "Message non reconnu"}))

async def start_game():
    p1 = players_queue.pop(0)
    p2 = players_queue.pop(0)

    logging.info(f"Partie commencée entre {p1['pseudo']} et {p2['pseudo']}")

    game_id = f"{p1['pseudo']}_vs_{p2['pseudo']}"
    board = [""]* 9  # Initialize a 3x3 board

    game[game_id] = {
        "player1" : p1,
        "player2" : p2,
        "board": board,
        "turn" : p1["pseudo"],  # Player 1 starts
    }

    await p1["ws"].send(json.dumps({
        "type": "game_start",
        "game_id": game_id,
        "opponent": p2["pseudo"],
        "symbol": "X",
        "your_turn": True
    }))
    await p2["ws"].send(json.dumps({
        "type": "game_start",
        "game_id": game_id,
        "opponent": p1["pseudo"],
        "symbol": "O",
        "your_turn": False
    }))

async def main():
    logging.info("Début du serveur WebSocket sur ws://localhost:8081")
    async with websockets.serve(handle_player, "0.0.0.0", 8081):
        await asyncio.Future() # run forever

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.info("Serveur WebSocket arrêté par l'utilisateur")


