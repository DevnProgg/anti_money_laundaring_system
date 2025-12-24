## Indexed Fields and Rationale

Here are the fields that should be indexed and the reasons why:

-   `account_id`: This field will be frequently used to filter transactions for a specific account. Creating an index on this column will significantly improve the performance of queries that retrieve the transaction history for a given account.
-   `timestamp`: Time-based queries are common in financial systems, for example, to retrieve transactions within a specific date range. An index on the timestamp column will speed up these queries.
-   `amount`: Indexing the amount field can be beneficial for queries that filter or sort transactions based on their value, such as finding all transactions above a certain threshold.
