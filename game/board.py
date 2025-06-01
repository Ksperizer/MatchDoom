import pygame 

WIDTH , HEIGHT = 800, 800
ROWS, COLS = 3, 3
CELL_SIZE = WIDTH // COLS
LINE_WIDTH = 15

WHITE = (255, 255, 255)
RED = (255, 50, 50)
BLUE = (0, 150, 255)
BLACK = (0, 0, 0)

class Board:
    def __init__ (self):
        self.grid = [["" for _ in range(COLS)] for _ in range(ROWS)]

    def draw(self, screen, font): 
        for i in range(1, ROWS):
            pygame.draw.line(screen, BLACK, (0, i * CELL_SIZE), (WIDTH, i * CELL_SIZE), LINE_WIDTH)
            pygame.draw.line(screen, BLACK, (i * CELL_SIZE, 0), (i * CELL_SIZE, HEIGHT), LINE_WIDTH)    

            for row in range(ROWS):
                for col in range(COLS):
                    if self.grid[row][col] != "":
                        symbol = font.render(self.grid[row][col], True, BLUE if self.grid[row][col] == "X" else RED)
                        rect = symbol.get_rect(center=(col * CELL_SIZE + CELL_SIZE  // 2, row * CELL_SIZE + CELL_SIZE // 2))
                        screen.blit(symbol, rect)
    

    def reset(self):
        self.grid = [[ "" for _ in range(COLS)] for _ in range(ROWS)]

    def is_full(self):
        return all(cell != "" for row in self.grid for cell in row)
    
    def set_cell(self, row, col, symbol):
        if self.grid[row][col] == "":
            self.grid[row][col] = symbol
            return True
        return False

    def check_winner(self):
        g = self.grid
        for i in range(ROWS):
            if g[i][0] == g[i][1] == g[i][2] != "":
                return g[i][0]
            if g[0][i] == g[1][i] == g[2][i] != "":
                return g[0][i]
        if g[0][0] == g[1][1] == g[2][2] != "":
            return g[0][0] 
        if g[0][2] == g[1][1] == g[2][0] != "":
            return g[0][2]
        if self.is_full():
            return "Draw"
        return None
    