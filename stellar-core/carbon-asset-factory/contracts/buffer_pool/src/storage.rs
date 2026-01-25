use soroban_sdk::{contracttype, Address, String};

#[contracttype]
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct CustodyRecord {
    pub token_id: u32,
    pub deposited_at: u64,
    pub depositor: Address,
    pub project_id: String,
}
