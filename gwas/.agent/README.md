# GWAS Database

Despite being named "GWAS Database", this database does more than host the GWAS Catalog.

The `gwas.duckdb` DuckDB database contains all the data required by various tools in the PHITE repo.

The `gwas.duckdb` is a local developer database, able to be created and recreated at will via `build_db.sh`. The build script is an entry point for all setup tasks to create the fully functional DuckDB database.

For any .agent "brief" (A markdown document expressing a feature or change), assume that any script or tool produced can be run on a new computer. Base items like python, etc. can be assumed to exist.

Note: It is expected that the build_db.sh will be the only script needed to run to install and prepare any required data.