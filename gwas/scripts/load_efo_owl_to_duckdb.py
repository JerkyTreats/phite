import rdflib
import duckdb
import os
import logging
from rdflib.namespace import RDF, RDFS, OWL

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(levelname)s %(message)s')
logger = logging.getLogger("efo_loader")

EFO_NS = rdflib.Namespace("http://www.ebi.ac.uk/efo/")
OBOINOWL = rdflib.Namespace("http://www.geneontology.org/formats/oboInOwl#")

OWL_FILE = "efo.owl"
DB_FILE = "gwas.duckdb"

def parse_efo_owl(owl_path):
    logger.info(f"Parsing OWL file: {owl_path}")
    import rdflib
    g = rdflib.Graph()
    import time
    start = time.time()
    g.parse(owl_path, format="xml")
    elapsed = time.time() - start
    logger.info(f"OWL file parsed in {elapsed:.2f} seconds.")
    logger.info("object parsed")
    terms = []
    synonyms = []
    relationships = []
    for s in g.subjects(RDF.type, OWL.Class):
        if not isinstance(s, rdflib.term.BNode):
            term_id = str(s)
            label = g.value(s, RDFS.label)
            definition = g.value(s, OBOINOWL.hasDefinition)
            terms.append({
                'uri': term_id,
                'label': str(label) if label else None,
                'definition': str(definition) if definition else None
            })
            # Synonyms
            for syn in g.objects(s, OBOINOWL.hasExactSynonym):
                synonyms.append({'uri': term_id, 'synonym': str(syn)})
            # Relationships (subclass, etc.)
            for o in g.objects(s, RDFS.subClassOf):
                if not isinstance(o, rdflib.term.BNode):
                    relationships.append({'subject': term_id, 'predicate': 'subClassOf', 'object': str(o)})
    logger.info(f"Parsed {len(terms)} terms, {len(synonyms)} synonyms, {len(relationships)} relationships.")
    return terms, synonyms, relationships

def create_schema(con):
    logger.info("Creating database schema (tables)...")
    con.execute("""
        CREATE TABLE IF NOT EXISTS efo_term (
            uri VARCHAR PRIMARY KEY,
            label VARCHAR,
            definition VARCHAR
        );
    """)
    con.execute("""
        CREATE TABLE IF NOT EXISTS efo_synonym (
            uri VARCHAR,
            synonym VARCHAR
        );
    """)
    con.execute("""
        CREATE TABLE IF NOT EXISTS efo_relationship (
            subject VARCHAR,
            predicate VARCHAR,
            object VARCHAR
        );
    """)
    logger.info("Schema created (if not already present).")

def insert_data(con, terms, synonyms, relationships):
    logger.info(f"Inserting data: {len(terms)} terms, {len(synonyms)} synonyms, {len(relationships)} relationships...")
    con.executemany("INSERT OR REPLACE INTO efo_term (uri, label, definition) VALUES (?, ?, ?)",
                    [(t['uri'], t['label'], t['definition']) for t in terms])
    if synonyms:
        con.executemany("INSERT INTO efo_synonym (uri, synonym) VALUES (?, ?)",
                        [(s['uri'], s['synonym']) for s in synonyms])
    if relationships:
        con.executemany("INSERT INTO efo_relationship (subject, predicate, object) VALUES (?, ?, ?)",
                        [(r['subject'], r['predicate'], r['object']) for r in relationships])

def main():
    import duckdb
    if not os.path.exists(OWL_FILE):
        logger.error(f"{OWL_FILE} not found")
        raise FileNotFoundError(f"{OWL_FILE} not found")
    logger.info(f"Parsing EFO OWL file: {OWL_FILE}")
    terms, synonyms, relationships = parse_efo_owl(OWL_FILE)
    logger.info(f"Parsed EFO OWL file: {len(terms)} terms, {len(synonyms)} synonyms, {len(relationships)} relationships.")
    con = duckdb.connect(DB_FILE)
    create_schema(con)
    logger.info(f"Inserting data into DuckDB: {DB_FILE}")
    insert_data(con, terms, synonyms, relationships)
    logger.info(f"Inserted")
    con.close()
    logger.info(f"Loaded {len(terms)} terms, {len(synonyms)} synonyms, {len(relationships)} relationships into {DB_FILE}")

if __name__ == "__main__":
    main()
