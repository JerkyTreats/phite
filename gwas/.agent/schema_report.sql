ATTACH 'hp.db' AS hp;

-- prefix
DESCRIBE hp.prefix;
SELECT * FROM hp.prefix LIMIT 5;

-- rdf_list_statement
DESCRIBE hp.rdf_list_statement;
SELECT * FROM hp.rdf_list_statement LIMIT 5;

-- rdf_level_summary_statistic
DESCRIBE hp.rdf_level_summary_statistic;
SELECT * FROM hp.rdf_level_summary_statistic LIMIT 5;

-- anonymous_expression
DESCRIBE hp.anonymous_expression;
SELECT * FROM hp.anonymous_expression LIMIT 5;

-- anonymous_class_expression
DESCRIBE hp.anonymous_class_expression;
SELECT * FROM hp.anonymous_class_expression LIMIT 5;

-- anonymous_property_expression
DESCRIBE hp.anonymous_property_expression;
SELECT * FROM hp.anonymous_property_expression LIMIT 5;

-- anonymous_individual_expression
DESCRIBE hp.anonymous_individual_expression;
SELECT * FROM hp.anonymous_individual_expression LIMIT 5;

-- owl_restriction
DESCRIBE hp.owl_restriction;
SELECT * FROM hp.owl_restriction LIMIT 5;

-- owl_complex_axiom
DESCRIBE hp.owl_complex_axiom;
SELECT * FROM hp.owl_complex_axiom LIMIT 5;

-- has_oio_synonym_statement
DESCRIBE hp.has_oio_synonym_statement;
SELECT * FROM hp.has_oio_synonym_statement LIMIT 5;

-- repair_action
DESCRIBE hp.repair_action;
SELECT * FROM hp.repair_action LIMIT 5;

-- problem
DESCRIBE hp.problem;
SELECT * FROM hp.problem LIMIT 5;

-- lexical_problem
DESCRIBE hp.lexical_problem;
SELECT * FROM hp.lexical_problem LIMIT 5;

-- relation_graph_construct
DESCRIBE hp.relation_graph_construct;
SELECT * FROM hp.relation_graph_construct LIMIT 5;

-- subgraph_query
DESCRIBE hp.subgraph_query;
SELECT * FROM hp.subgraph_query LIMIT 5;

-- entailed_edge
DESCRIBE hp.entailed_edge;
SELECT * FROM hp.entailed_edge LIMIT 5;

-- term_association
DESCRIBE hp.term_association;
SELECT * FROM hp.term_association LIMIT 5;

-- statements
DESCRIBE hp.statements;
SELECT * FROM hp.statements LIMIT 5;
