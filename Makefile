NAME		=	concorrencia

DB_FILES	=	*.db *.db-wal *.db-shm

.PHONY: all clean fclean re stop

all:
	go build -o $(NAME)
	@./$(NAME) &
	@sleep 1
	@echo ""
	@echo "  ========================================="
	@echo "  Servidor rodando em background"
	@echo "  ========================================="
	@echo ""
	@echo "  Tela 1 (Load Test)  → http://localhost:8080/"
	@echo "  Tela 2 (Saldo)      → http://localhost:8080/saldo.html"
	@echo ""
	@echo "  PID: $$(pgrep -f './$(NAME)' | head -1)"
	@echo ""
	@echo "  Para parar: make stop"
	@echo "  ========================================="
	@echo ""

clean:
	rm -f $(NAME)

fclean: clean
	rm -f $(DB_FILES)

re: fclean all

stop:
	-pkill -f './$(NAME)'
