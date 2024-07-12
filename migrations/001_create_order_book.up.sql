CREATE TABLE IF NOT EXISTS OrderBook (
    id Int64,
    exchange String,
    pair String,
    asks Array(Tuple(Float64, Float64)),
    bids Array(Tuple(Float64, Float64))
) ENGINE = MergeTree() ORDER BY id;
