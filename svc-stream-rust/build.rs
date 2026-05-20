fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Corrigido para compile_protos e atualizado o caminho conforme a nova árvore
    tonic_build::compile_protos("../contracts/cambio/regulatorio/v1/cambio.proto")?;
    Ok(())
}

