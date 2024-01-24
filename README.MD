## This is a simple project to learn Golang and Docker

##### Project is a simple API to tracking pricing and alert price of product on Shopee e-commerce platform

#### Why this project is stopped?

##### Because Shopee always change strategy to prevent scraping data from their website.

##### So, I can not spend more time to find a way to scrape data from Shopee website.

##### I will continue this project when I find a way to scrape data from Shopee website.

### How to run this project?

```bash
// Build and run
docker compose up -d

// Some command docker for new guy
docker compose exec api bash
or
docker compose run --service-ports api bash
docker compose build
docker compose up -d
docker compose logs -f api
docker compose stop
docker-compose down --rmi all --volumes
docker compose ps
docker compose images
```
