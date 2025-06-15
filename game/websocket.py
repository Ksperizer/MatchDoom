import asyncio 
import websockets
import json
import logging

logging.basicConfig(level=logging.INFO)

clients = set()

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
    
async def handle_message(data, websocket):
    msg_type = data.get("type")

    if msg_type == "ping":
        return {"type": "pong"}
    elif msg_type == "join":
        pseudo = data.get("pseudo")
        logging.info(f"Pseudo {pseudo} a rejoint la file d'attente")
        return {
            "type": "queue",
            "message": "En attente d'un adversaire",
            "position": len(clients)
        }
    elif msg_type == "test":
        return {"type": "info", "message":  "Serveur python en ligne "}
    else:
        return {"type": "error", "message": "message non reconnu"}

async def main():
    logging.info("Debut du serveur WebSocket sur ws://localhost:8081")
    async with websockets.serve(handle_player, "0.0.0.0", 8081):
        await asyncio.Future() # run forever

if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        logging.info("Serveur WebSocket arrêté par l'utilisateur")


