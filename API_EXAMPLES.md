
## Example `curl` Command

Here is an example `curl` command to test the `/transactions` endpoint:

```bash
curl -X POST http://localhost:8080/transactions \
-H "Content-Type: application/json" \
-d '{ \
    "account_id": "acc-123", \
    "amount": 100.50, \
    "currency": "USD", \
    "source_country": "US", \
    "destination_country": "CA", \
    "transaction_type": "transfer", \
    "status": "pending" \
}'
```

## Expected Responses

### Success Response (Status 201 Created)

```json
{
    "transaction_id": "a8c7b6a5-4f3d-4e2a-8b1e-9e6a7c5d4b3a"
}
```

### Error Response (Status 400 Bad Request)

This response is returned when the request body is invalid or fails validation.

```
amount must be greater than 0
```

### Error Response (Status 500 Internal Server Error)

This response is returned when there is a database error.

```
Failed to create transaction
```
