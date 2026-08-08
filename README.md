# go-todo-api

### Database

```docker run --name todo-postgres 
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=todo_db \
  -p 5432:5432 \
  -d postgres:16
```

- Connecting DB: #Connect: docker exec -it todo-postgres psql -U postgres -d todo_db

