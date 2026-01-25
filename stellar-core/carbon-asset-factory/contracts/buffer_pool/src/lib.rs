#![no_std]

mod errors;
mod storage;

use soroban_sdk::{contract, contractimpl, Env};

#[contract]
pub struct BufferPoolContract;

#[contractimpl]
impl BufferPoolContract {
    pub fn placeholder(_env: Env) -> u32 {
        0
    }
}
