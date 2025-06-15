import json
import threading
import websocket

class SocketClient: 
    def __init__(self, url="ws://localhost:8081/game/ws", pseudo="PlayerX"):
        self.url = url
        self.pseudo = pseudo
        self.ws = websocket.WebSocketApp(
            self.url, 
            on_open=self.on_open,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close
        )
        self.thread = threading.Thread(target=self.ws.run_forever)
        self.thread.deamon = True
        self.thread.start()
        # self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        # self.sock.connect((host, port))

    def on_open(self, ws):
        print("Connexion WebSocket ouverte")
        # Send a message 'join'
        ws.send(json.dumps({
            "type": "join",
            "pseudo": self.pseudo
        }))
        
    def on_message(self, ws, message):
        print("Message reçu:", message)
        # Handle incoming messages here
        data = json.loads(message)
        if data["type"]== "game_start":
           print(f"Partie contre {data['opponent']} a commencé !")
        elif data["type"] == "move_played":
            print(f"{data['player']} a joué à la position {data['row']}, {data['col']}")


                 

    def send_move(self, row, col):
        message = json.dumps({
            "type": "move",
            "row": row,
            "col": col
        })
        self.sock.sendall(message.encode())
    
    def receive(self):
        try: 
            data = self.sock.recv(1024)
            return data.decode()
        except:
            return None
        
    def close(self):
        self.sock.close()
