use reth_db::open_db_read_only;
use reth_db_provider::BlockReader;
use reth_db_provider::{
    BlockExecutor, ProviderFactory
};

use reth_primitives::{
    MAINNET, BlockHashOrNumber, U256
};

use reth_revm::{
    database::StateProviderDatabase,
    processor::EVMProcessor
};

use std::path::Path;
use std::sync::Arc;
use rand::distributions::{Distribution, Uniform};
use tracing::info;
use std::time::SystemTime;

// #[derive(Parser, Debug)]

fn main() {
    let mut cur_time = SystemTime::now();
    println!("Start Current Time is {:?}", cur_time);

    let string = String::from("/home/user/common/docker/volumes/eth-docker_reth-el-data/_data/db");
    let path = Path::new(&string);
    let db = Arc::new(open_db_read_only(&path, None).unwrap());
    let chain_spec = MAINNET.clone();

    let provider = Arc::new(ProviderFactory::new(&db, chain_spec.clone()));

    let loop_time = 100;

    for i in 0..loop_time {
        // let block_range = 18650000..18768100;

        // let mut rng = rand::thread_rng();
        // let block_uniform_range = Uniform::from(block_range);

        // let old_block_num= block_uniform_range.sample(&mut rng);
        let old_block_num = 18768099;
        let new_block_num = old_block_num + 1;
        let new_block = provider.block(BlockHashOrNumber::Number(new_block_num)).unwrap().unwrap();


        let state_provider = provider.history_by_block_number(old_block_num).unwrap();

        let mut executor = EVMProcessor::new_with_db(chain_spec.clone(), StateProviderDatabase::new(state_provider));

        // let result = executor.execute_and_verify_receipt(&new_block, U256::ZERO, None).unwrap();

        let result = executor.execute(&new_block, U256::ZERO, None).unwrap();

        let stat = executor.stats();
        println!("Show stats: {:?}", stat);

        println!("new block number {:?}, round: {:?}", new_block_num, i);
    }
    cur_time = SystemTime::now();
    println!("End Current Time is {:?}", cur_time);

}
