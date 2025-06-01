-- Reference Panel Table for PRS Normalization
-- Stores full, unfiltered sample metadata from gnomAD v3.1.2
-- Schema: all columns from gnomAD v3.1.2 sample panel metadata

CREATE TABLE IF NOT EXISTS reference_panel (
    s TEXT,
    bam_metrics TEXT,
    sample_qc TEXT,
    gnomad_sex_imputation TEXT,
    gnomad_population_inference TEXT,
    gnomad_sample_qc_residuals TEXT,
    gnomad_sample_filters TEXT,
    gnomad_high_quality TEXT,
    gnomad_release TEXT,
    relatedness_inference TEXT,
    hgdp_tgp_meta TEXT,
    high_quality TEXT
);
