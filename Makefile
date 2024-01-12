project_name = wallet_service

.docker-build:
	docker compose -f $(project_name).yml -p $(project_name) up --build