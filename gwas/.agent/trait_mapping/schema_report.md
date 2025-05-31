# GWAS Database Schema Report

This report describes tables in `gwas.duckdb` (with `hp.db` attached) and provides schema and sample data for each table listed.

---

## `prefix`

### Table Description

```
column_name | column_type | null | key | default | extra
prefix      | VARCHAR     | YES  | NULL| NULL    | NULL
base        | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
prefix  | base
rdf     | http://www.w3.org/1999/02/22-rdf-syntax-ns#
rdfs    | http://www.w3.org/2000/01/rdf-schema#
xsd     | http://www.w3.org/2001/XMLSchema#
owl     | http://www.w3.org/2002/07/owl#
```

## `rdf_list_statement`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
value       | VARCHAR     | YES  | NULL| NULL    | NULL
datatype    | VARCHAR     | YES  | NULL| NULL    | NULL
language    | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `rdf_level_summary_statistic`

### Table Description

```
column_name | column_type | null | key | default | extra
element     | VARCHAR     | YES  | NULL| NULL    | NULL
count_value | BIGINT      | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `anonymous_expression`

### Table Description

```
column_name | column_type | null | key | default | extra
id          | VARCHAR     | NO   | PRI | NULL    | NULL
```

### Table Example

```
(no rows)
```

## `anonymous_class_expression`

### Table Description

```
column_name | column_type | null | key | default | extra
id          | VARCHAR     | NO   | PRI | NULL    | NULL
```

### Table Example

```
(no rows)
```

## `anonymous_property_expression`

### Table Description

```
column_name | column_type | null | key | default | extra
id          | VARCHAR     | NO   | PRI | NULL    | NULL
```

### Table Example

```
(no rows)
```

## `anonymous_individual_expression`

### Table Description

```
column_name | column_type | null | key | default | extra
id          | VARCHAR     | NO   | PRI | NULL    | NULL
```

### Table Example

```
(no rows)
```

## `owl_restriction`

### Table Description

```
column_name | column_type | null | key | default | extra
on_property | VARCHAR     | YES  | NULL| NULL    | NULL
filler      | VARCHAR     | YES  | NULL| NULL    | NULL
id          | VARCHAR     | NO   | PRI | NULL    | NULL
```

### Table Example

```
(no rows)
```

## `owl_complex_axiom`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `has_oio_synonym_statement`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
value       | VARCHAR     | NO   | NULL| NULL    | NULL
datatype    | VARCHAR     | YES  | NULL| NULL    | NULL
language    | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `repair_action`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
description | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `problem`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
value       | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `lexical_problem`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
value       | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `relation_graph_construct`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `subgraph_query`

### Table Description

```
column_name | column_type | null | key | default | extra
subject          | VARCHAR     | YES  | NULL| NULL    | NULL
predicate        | VARCHAR     | YES  | NULL| NULL    | NULL
object           | VARCHAR     | YES  | NULL| NULL    | NULL
anchor_object    | VARCHAR     | YES  | NULL| NULL    | NULL
anchor_predicate | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `entailed_edge`

### Table Description

```
column_name | column_type | null | key | default | extra
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
subject     | predicate        | object
HP:0000679  | rdfs:subClassOf  | HP:0000679
HP:0008458  | rdfs:subClassOf  | HP:0008458
HP:0011402  | rdfs:subClassOf  | HP:0011402
HP:0100158  | rdfs:subClassOf  | HP:0100158
GO:0061047  | rdfs:subClassOf  | GO:0061047
```

## `term_association`

### Table Description

```
column_name | column_type | null | key | default | extra
id            | VARCHAR     | NO   | PRI | NULL    | NULL
subject       | VARCHAR     | YES  | NULL| NULL    | NULL
predicate     | VARCHAR     | YES  | NULL| NULL    | NULL
object        | VARCHAR     | YES  | NULL| NULL    | NULL
evidence_type | VARCHAR     | YES  | NULL| NULL    | NULL
publication   | VARCHAR     | YES  | NULL| NULL    | NULL
source        | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
(no rows)
```

## `statements`

### Table Description

```
column_name | column_type | null | key | default | extra
stanza      | VARCHAR     | YES  | NULL| NULL    | NULL
subject     | VARCHAR     | YES  | NULL| NULL    | NULL
predicate   | VARCHAR     | YES  | NULL| NULL    | NULL
object      | VARCHAR     | YES  | NULL| NULL    | NULL
value       | VARCHAR     | YES  | NULL| NULL    | NULL
datatype    | VARCHAR     | YES  | NULL| NULL    | NULL
language    | VARCHAR     | YES  | NULL| NULL    | NULL
graph       | VARCHAR     | YES  | NULL| NULL    | NULL
```

### Table Example

```
stanza      | subject    | predicate                          | object | value
obo:hp.owl | obo:hp.owl | rdfs:comment                       | NULL   | Please see license of HPO at http://www.human-phenotype-ontology.org
obo:hp.owl | obo:hp.owl | oio:logical-definition-view-relation| NULL   | has_part
obo:hp.owl | obo:hp.owl | oio:default-namespace              | NULL   | human_phenotype
obo:hp.owl | obo:hp.owl | dcterms:license                    | <https://hpo.jax.org/app/license> | NULL
obo:hp.owl | obo:hp.owl | dce:title                          | NULL   | Human Phenotype Ontology
```
