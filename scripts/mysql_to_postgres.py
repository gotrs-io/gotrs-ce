#!/usr/bin/env python3
"""
Convert MySQL schema to PostgreSQL
"""

import re
import sys

def convert_mysql_to_postgres(mysql_sql):
    """Convert MySQL CREATE TABLE statements to PostgreSQL"""
    
    lines = mysql_sql.split('\n')
    output = []
    in_table = False
    columns = []
    constraints = []
    indexes = []
    
    for line in lines:
        line = line.strip()
        
        # Start of CREATE TABLE
        if line.startswith('CREATE TABLE'):
            in_table = True
            table_name = re.search(r'CREATE TABLE `?(\w+)`?', line).group(1)
            output.append(f'CREATE TABLE {table_name} (')
            columns = []
            constraints = []
            indexes = []
            continue
            
        # End of CREATE TABLE
        if in_table and line.startswith(') ENGINE='):
            # Add columns
            all_items = columns + [c for c in constraints if not c.startswith('CONSTRAINT FK_')]
            
            for i, item in enumerate(all_items):
                if i < len(all_items) - 1:
                    output.append(f'  {item},')
                else:
                    output.append(f'  {item}')
            
            output.append(');')
            output.append('')
            
            # Add indexes after table
            for idx in indexes:
                if not idx.startswith('-- FK index'):
                    output.append(idx)
            
            in_table = False
            continue
            
        # Process columns and constraints
        if in_table:
            # Remove backticks
            line = line.replace('`', '')
            
            # Column definition
            if re.match(r'^\w+\s+', line) and not line.startswith('PRIMARY') and not line.startswith('KEY') and not line.startswith('UNIQUE') and not line.startswith('CONSTRAINT'):
                # Convert data types
                col_line = line.rstrip(',')
                
                # Handle AUTO_INCREMENT
                if 'AUTO_INCREMENT' in col_line:
                    if 'bigint' in col_line.lower():
                        col_line = re.sub(r'bigint\(\d+\)\s+NOT NULL AUTO_INCREMENT', 'BIGSERIAL PRIMARY KEY', col_line)
                    elif 'smallint' in col_line.lower():
                        col_line = re.sub(r'smallint\(\d+\)\s+NOT NULL AUTO_INCREMENT', 'SMALLSERIAL PRIMARY KEY', col_line)
                    else:
                        col_line = re.sub(r'int\(\d+\)\s+NOT NULL AUTO_INCREMENT', 'SERIAL PRIMARY KEY', col_line)
                else:
                    # Convert integer types
                    col_line = re.sub(r'int\(\d+\)', 'INTEGER', col_line)
                    col_line = re.sub(r'bigint\(\d+\)', 'BIGINT', col_line)
                    col_line = re.sub(r'smallint\(\d+\)', 'SMALLINT', col_line)
                    col_line = re.sub(r'tinyint\(\d+\)', 'SMALLINT', col_line)
                    
                # Convert other types
                col_line = col_line.replace('datetime', 'TIMESTAMP')
                col_line = col_line.replace('longtext', 'TEXT')
                col_line = col_line.replace('mediumtext', 'TEXT')
                col_line = col_line.replace('longblob', 'BYTEA')
                col_line = col_line.replace('mediumblob', 'BYTEA')
                col_line = col_line.replace('blob', 'BYTEA')
                
                # Fix capitalization issues
                col_line = col_line.replace('smallINTEGER', 'SMALLINT')
                col_line = col_line.replace('bigINTEGER', 'BIGINT')
                col_line = col_line.replace('smallSERIAL', 'SMALLSERIAL')
                
                # Handle text type with no size
                col_line = re.sub(r'\btext\b', 'TEXT', col_line)
                
                columns.append(col_line)
                
            # PRIMARY KEY
            elif line.startswith('PRIMARY KEY'):
                # Already handled in column if AUTO_INCREMENT
                if not any('SERIAL' in col for col in columns):
                    pk_col = re.search(r'PRIMARY KEY \(([^)]+)\)', line).group(1)
                    constraints.append(f'PRIMARY KEY ({pk_col})')
                    
            # UNIQUE KEY
            elif line.startswith('UNIQUE KEY'):
                match = re.search(r'UNIQUE KEY (\w+) \(([^)]+)\)', line)
                if match:
                    idx_name = match.group(1)
                    idx_cols = match.group(2)
                    constraints.append(f'UNIQUE ({idx_cols})')
                    
            # Regular KEY (index)
            elif line.startswith('KEY'):
                match = re.search(r'KEY (\w+) \(([^)]+)\)', line)
                if match:
                    idx_name = match.group(1)
                    idx_cols = match.group(2)
                    # Remove MySQL-specific length specifications in indexes
                    idx_cols = re.sub(r'\(\d+\)', '', idx_cols)
                    # Skip FK indexes, they're implied
                    if not idx_name.startswith('FK_'):
                        indexes.append(f'CREATE INDEX {idx_name} ON {table_name} ({idx_cols});')
                        
            # CONSTRAINT (foreign keys) - skip for baseline, add separately if needed
            elif line.startswith('CONSTRAINT'):
                pass
    
    return '\n'.join(output)

if __name__ == '__main__':
    # Read MySQL schema
    with open('schema/baseline/otrs_mysql_structure.sql', 'r') as f:
        mysql_sql = f.read()
    
    # Convert to PostgreSQL
    postgres_sql = convert_mysql_to_postgres(mysql_sql)
    
    # Write PostgreSQL schema
    with open('schema/baseline/otrs_complete.sql', 'w') as f:
        f.write(postgres_sql)
    
    print(f"Converted {mysql_sql.count('CREATE TABLE')} tables to PostgreSQL")