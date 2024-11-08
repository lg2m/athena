use std::path::Path;

use anyhow::Result;
use athena_term::run_editor;

#[tokio::main]
async fn main() -> Result<()> {
    run_editor().await
}
