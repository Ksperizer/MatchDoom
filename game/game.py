import pygame
import sys
import socket
from game.board import WIDTH, HEIGHT,CELL_SIZE, Board, WHITE, BLACK
from game.morpion import SocketClient

class Game:
    def __init__(self):
        pygame.init()
        self.screen = pygame.display.set_mode((WIDTH, HEIGHT))
        pygame.display.set_caption("Morpion - MatchDoom")
        self.font = pygame.font.SysFont(None, 120)
        self.small_font = pygame.font.SysFont(None, 50)
        self.board = Board()
        self.current_player = "X"
        self.running = True
        self.game_over = False
        self.client = SocketClient()  

    def run(self):
        while self.running:
            self.handle_events()
            self.draw()
            pygame.display.flip()

    def handle_events(self):
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                self.running = False
                pygame.quit()
                sys.exit()

            if event.type == pygame.MOUSEBUTTONDOWN and not self.game_over:
                x, y = event.pos
                row, col = y // CELL_SIZE, x // CELL_SIZE
                if self.board.set_cell(row, col, self.current_player):
                    winner = self.board.check_winner()
                    if winner:
                        self.game_over = True
                        self.show_winner(winner)
                        pygame.time.delay(2000)
                        self.board.reset()
                        self.current_player = "X"
                        self.game_over = False
                    else:
                        self.current_player = "O" if self.current_player == "X" else "X"
        
        self.client.send_move(row, col)
        print("Move sent to server:", row, col)

    def draw(self):
        self.screen.fill(WHITE)
        self.board.draw(self.screen, self.font)

    def show_winner(self, winner):
        text = "Égalité !" if winner == "Draw" else f"{winner} a gagné !"
        render = self.small_font.render(text, True, BLACK)
        rect = render.get_rect(center=(WIDTH // 2, HEIGHT // 2))
        self.screen.blit(render, rect)
        pygame.display.flip()