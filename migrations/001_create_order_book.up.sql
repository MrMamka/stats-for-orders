CREATE TABLE IF NOT EXISTS OrderBook (
    id Int64,
    exchange String,
    pair String,
    asks Array(Array(Float64)),
    bids Array(Array(Float64))
) ENGINE = MergeTree() 
ORDER BY (exchange, pair);
