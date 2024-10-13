use anyhow::Result;

use athena_term::render::run_editor;

#[tokio::main]
async fn main() -> Result<()> {
    run_editor().await
}
