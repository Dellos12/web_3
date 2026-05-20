
-- Ativa a extensão de álgebra linear e vetores na versão instalada localmente
CREATE EXTENSION IF NOT EXISTS vector;

-- Tabela para armazenar o espaço vetorial regulatório (CNAEs e Regras do IVA Dual)
CREATE TABLE IF NOT EXISTS matriz_regulatoria (
    id SERIAL PRIMARY KEY,
    cnae_codigo VARCHAR(10) NOT NULL,
    descricao_regra TEXT NOT NULL,
    aliquota_cbs_base NUMERIC(5,2) NOT NULL, -- CBS Federal
    aliquota_ibs_base NUMERIC(5,2) NOT NULL, -- IBS Estadual/Municipal
    
    -- O vetor correspondente à norma jurídica (ex: 1536 dimensões do padrão de mercado)
    regra_embedding vector(1536) NOT NULL,
    
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Tabela de Logs de Auditoria onde o SHA é cruzado com a decisão da IA
CREATE TABLE IF NOT EXISTS auditoria_operacoes (
    transacao_id UUID PRIMARY KEY,
    documento_fiscal_sha256 CHAR(64) NOT NULL UNIQUE, -- Lacre Criptográfico
    cnae_processado VARCHAR(10) NOT NULL,
    valor_brl NUMERIC(15,2) NOT NULL,
    cosseno_similaridade_calculado DOUBLE PRECISION NOT NULL,
    rail_escolhido VARCHAR(30) NOT NULL,
    hash_auditoria_estado CHAR(64) NOT NULL, -- Assinatura gravada na Web3
    data_processamento TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- A REGRA DE OURO DA PERFORMANCE: Criação do Índice HNSW
-- Usamos 'vector_cosine_ops' para que o banco calcule o Cosseno em velocidade de hardware
CREATE INDEX IF NOT EXISTS idx_matriz_regulatoria_hnsw 
ON matriz_regulatoria 
USING hnsw (regra_embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);

-- Massa de dados fictícia para teste do nosso canteiro de obras
-- Inserindo um vetor normalizado simulado para o CNAE de Dropshipping B2B com IVA Dual de 2026
INSERT INTO matriz_regulatoria (cnae_codigo, descricao_regra, aliquota_cbs_base, aliquota_ibs_base, regra_embedding)
VALUES (
    '7490-1/04', -- Exemplo: Atividades de intermediação de negócios (Dropshipping)
    'Regra Geral de Importação B2B - Intermediação por Princípio do Destino com Split Payment',
    8.80,  -- CBS simulada
    17.70, -- IBS simulado (Total 26.5% IVA estimado para 2026)
    -- Simulando o início de um vetor de 1536 dimensões preenchido com valores de exemplo
    (SELECT array_agg(0.025::float)::vector FROM generate_series(1, 1536))
) ON CONFLICT DO NOTHING;
