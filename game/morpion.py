import socket 
import json 

class SocketClient: 
    def __init__(self, host='localhost', port=9000):
        self.sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self.sock.connect((host, port))

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

    