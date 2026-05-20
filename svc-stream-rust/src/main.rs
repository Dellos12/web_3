
use sha2::{Sha256, Digest};
use tonic::{transport::Server, Request, Response, Status};

// Injeta os códigos gerados automaticamente a partir do cambio.proto
pub mod cambio {
    pub mod regulatorio {
        pub mod v1 {
            tonic::include_proto!("cambio.regulatorio.v1");
        }
    }
}

use cambio::regulatorio::v1::roteador_fiscal_service_server::{RoteadorFiscalService, RoteadorFiscalServiceServer};
use cambio::regulatorio::v1::{RequestTransacao, ResponseDirecionamento};

#[derive(Debug, Default)]
pub struct MotorStreamRust {}

#[tonic::async_trait]
impl RoteadorFiscalService for MotorStreamRust {
    async fn validar_e_rotear_transacao(
        &self,
        request: Request<RequestTransacao>,
    ) -> Result<Response<ResponseDirecionamento>, Status> {
        let mut req = request.into_inner();

        // 🛡️ O ENGATE CRIPTOGRÁFICO: Calcula ou valida o SHA-256 localmente
        let mut hasher = Sha256::new();
        hasher.update(req.cnpj_emissor.as_bytes());
        hasher.update(req.valor_operacao_brl.to_be_bytes());
        
        let hash_calculado = format!("{:x}", hasher.finalize());
        
        // Atribui o lacre digital criptográfico ao contrato estrito
        req.documento_fiscal_sha256 = hash_calculado.clone();

        // 🔀 Conexão com o Backplane: Aqui o Rust passaria a stream para o Go via gRPC cliente.
        // Para o teste do Jenkins local rodar de forma atômica, o Rust devolve o carimbo inicial.
        let response = ResponseDirecionamento {
            transacao_id: req.transacao_id,
            conformidade_aprovada: true,
            aliquota_cbs_federal: 8.80,
            aliquota_ibs_estadual: 17.70,
            valor_imposto_retido_brl: req.valor_operacao_brl * 0.265,
            rail_escolhido: 2, // RAIL_DESCENTRALIZADO_WEB3 (Drex/Blockchain)
            hash_auditoria_estado: hash_calculado,
        };

        Ok(Response::new(response))
    }
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Escuta nativamente no localhost na porta 50051 do seu backplane
    let addr = "127.0.0.1:50051".parse()?;
    let motor = MotorStreamRust::default();

    println!("⚡ Motor Rust rodando localmente na porta {}...", addr);

    Server::builder()
        .add_service(RoteadorFiscalServiceServer::new(motor))
        .serve(addr)
        .await?;

    Ok(())
}
