# Voucher Manager

## Architecture
Service will listen through HTTP API for voucher redemption requests. Redeeming is done atomically to avoid inconsistency. 

Any insertion will be checked with redis with atomic `Decr` command. allowed writes are then persisted into database. any failure on insert (i.e. duplicate record) will revert the `Decr` method.