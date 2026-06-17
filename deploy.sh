
set -euo pipefail

docker build -t nustdevs-api:latest .
docker stack deploy -c docker-compose.prod.yml nustdevs
docker service update --force nustdevs_api
docker service update --force nustdevs_worker
