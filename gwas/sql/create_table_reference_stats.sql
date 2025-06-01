-- Reference SNP Statistics Table for PRS Normalization
-- This table is required by the polygenic-risk-calculator for ancestry-aware normalization.
-- Generated from gnomAD v3.1.2 (GRCh38) reference panel VCFs and sample metadata.

CREATE TABLE IF NOT EXISTS reference_stats (
    rsid TEXT,           -- dbSNP ID (e.g., rs12345)
    ancestry TEXT,       -- Ancestry group (e.g., EUR)
    allele TEXT,         -- Effect/reference allele
    allele_freq DOUBLE,  -- Frequency of the effect/reference allele in the ancestry group
    mean_prs DOUBLE,     -- Mean PRS for ancestry group (optional/future)
    sd_prs DOUBLE,       -- SD of PRS for ancestry group (optional/future)
    PRIMARY KEY (rsid, ancestry, allele)
);

-- Example insert (to be replaced by automated ingestion):
-- INSERT INTO reference_stats (rsid, ancestry, allele, allele_freq) VALUES ('rs12345', 'EUR', 'A', 0.123);
