https://github.com/hashicorp/hcl
https://deepwiki.com/hashicorp/hcl

HCL Overview

This document provides a comprehensive overview of the HashiCorp Configuration Language (HCL) v2 codebase, its design philosophy, architecture, and core components. HCL is a toolkit for creating structured configuration languages that are both human- and machine-friendly.

For detailed information about specific subsystems, see Architecture and Components and Core Concepts. For expression evaluation specifics, see Expression System. For parsing and decoding mechanisms, see Parsing and Decoding.

Design Philosophy

HCL is designed as a configuration language toolkit rather than a single monolithic parser. It strikes a balance between human readability and machine processability, supporting both a native syntax optimized for human authoring and a JSON-based variant for programmatic generation.

The library is built around the principle of syntax-agnostic processing - the same API and processing pipeline works regardless of whether the input is HCL native syntax or JSON. This design allows applications to accept configuration in either format seamlessly.

Key Design Principles:

    Declarative Configuration: Focus on describing "what" rather than "how"
    Expression Support: Built-in expression evaluation for dynamic configuration
    Type Safety: Integration with the go-cty type system for robust value handling
    Modular Architecture: Separate concerns across focused packages
    Error Handling: Rich diagnostic information with precise source location tracking

Information Model

HCL's information model is built around two primary constructs: attributes and blocks. This model is syntax-agnostic and consistent across both native HCL and JSON representations.

Core Information Model

Attribute Example:

io_mode = "async"
count = 1 + 2

Block Example:

service "http" "web_proxy" {
  listen_addr = "127.0.0.1:8080"

  process "main" {
    command = ["/usr/local/bin/awesome-app", "server"]
  }
}

Package Architecture

HCL's modular architecture separates concerns across focused packages, each with specific responsibilities.
Package	Primary Interface	Responsibility
hcl	Body, Expression	Core interfaces and common types
hclsyntax	ParseConfig()	Native HCL syntax parsing and evaluation
json	Parse()	JSON syntax parsing with HCL semantics
hclparse	Parser.ParseHCL()	File-based parsing with caching
hcldec	Decode(spec, body, ctx)	Specification-based decoding
gohcl	DecodeBody()	Go struct-based decoding
hclwrite	NewFile(), Format()	Programmatic HCL generation

Extension Packages:

    ext/dynblock - Dynamic block generation with for_each
    ext/typeexpr - Type constraint expressions and validation
    ext/tryfunc - Error handling functions (try, can)
    ext/customdecode - Custom decoding behavior hooks

Integration with go-cty

HCL is tightly integrated with the go-cty library for its type system and value representation. This integration provides:

Type System Features:

    Strong typing with cty.Type constraints
    Unknown value handling for partial evaluation
    Value marking for sensitive data tracking
    Type refinements for more precise unknown values

Key Integration Points:

    hcl.Expression.Value() returns cty.Value
    hcl.EvalContext contains cty.Value variables and functions
    All decoding operations work with cty.Value as intermediate representation
    Type constraints in extensions use cty.Type

Major Use Cases

HCL supports several primary use cases through its modular architecture:

    Configuration File Parsing - Parse HCL/JSON files into structured data
    Dynamic Configuration - Evaluate expressions with variables and functions
    Code Generation - Programmatically generate well-formatted HCL
    Schema Validation - Validate configuration against predefined specifications
    Template Processing - Handle interpolation and control flow in templates

Each use case is supported by specific package combinations optimized for that workflow.
