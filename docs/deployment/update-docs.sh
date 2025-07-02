#!/bin/bash

# Documentation Update Script
# This script updates and validates the documentation for the WhatsApp Go library

set -e  # Exit on any error

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
DOCS_DIR="$PROJECT_ROOT/docs"
TIMESTAMP=$(date '+%Y-%m-%d')

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to validate markdown files
validate_markdown() {
    local file="$1"
    log_info "Validating markdown file: $file"
    
    # Check if file exists
    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        return 1
    fi
    
    # Check file size (should not be empty)
    if [[ ! -s "$file" ]]; then
        log_error "File is empty: $file"
        return 1
    fi
    
    # Basic markdown validation (check for common issues)
    if grep -q "]()" "$file"; then
        log_warning "Found empty links in: $file"
    fi
    
    log_success "Markdown validation passed: $file"
    return 0
}

# Function to validate JSON files
validate_json() {
    local file="$1"
    log_info "Validating JSON file: $file"
    
    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        return 1
    fi
    
    if command_exists jq; then
        if jq empty "$file" >/dev/null 2>&1; then
            log_success "JSON validation passed: $file"
            return 0
        else
            log_error "Invalid JSON syntax in: $file"
            return 1
        fi
    else
        log_warning "jq not found, skipping JSON validation for: $file"
        return 0
    fi
}

# Function to validate YAML files
validate_yaml() {
    local file="$1"
    log_info "Validating YAML file: $file"
    
    if [[ ! -f "$file" ]]; then
        log_error "File not found: $file"
        return 1
    fi
    
    if command_exists yq; then
        if yq eval '.' "$file" >/dev/null 2>&1; then
            log_success "YAML validation passed: $file"
            return 0
        else
            log_error "Invalid YAML syntax in: $file"
            return 1
        fi
    elif command_exists python3; then
        if python3 -c "import yaml; yaml.safe_load(open('$file'))" >/dev/null 2>&1; then
            log_success "YAML validation passed: $file"
            return 0
        else
            log_error "Invalid YAML syntax in: $file"
            return 1
        fi
    else
        log_warning "No YAML validator found, skipping validation for: $file"
        return 0
    fi
}

# Function to update timestamps in documentation
update_timestamps() {
    log_info "Updating timestamps in documentation files"
    
    local files=(
        "$DOCS_DIR/COMPLETE_GUIDE.md"
        "$DOCS_DIR/llm/whatsapp-go-docs.txt"
        "$DOCS_DIR/llm/whatsapp-go-schema.json"
        "$DOCS_DIR/llm/whatsapp-go-config.yaml"
    )
    
    for file in "${files[@]}"; do
        if [[ -f "$file" ]]; then
            # Update last updated date
            if [[ "$file" == *.md ]]; then
                sed -i.bak "s/\*This documentation was last updated on [0-9-]*\./\*This documentation was last updated on $TIMESTAMP\./g" "$file"
                rm -f "$file.bak"
            elif [[ "$file" == *.txt ]]; then
                sed -i.bak "s/Last Updated: [0-9-]*/Last Updated: $TIMESTAMP/g" "$file"
                rm -f "$file.bak"
            elif [[ "$file" == *.json ]]; then
                if command_exists jq; then
                    jq ".library.last_updated = \"$TIMESTAMP\"" "$file" > "$file.tmp" && mv "$file.tmp" "$file"
                fi
            elif [[ "$file" == *.yaml ]]; then
                sed -i.bak "s/last_updated: \"[0-9-]*\"/last_updated: \"$TIMESTAMP\"/g" "$file"
                rm -f "$file.bak"
            fi
            log_success "Updated timestamp in: $file"
        fi
    done
}

# Function to check documentation completeness
check_completeness() {
    log_info "Checking documentation completeness"
    
    local required_files=(
        "$DOCS_DIR/INDEX.md"
        "$DOCS_DIR/COMPLETE_GUIDE.md"
        "$DOCS_DIR/structured/README.md"
        "$DOCS_DIR/llm/whatsapp-go-docs.txt"
        "$DOCS_DIR/llm/whatsapp-go-schema.json"
        "$DOCS_DIR/llm/whatsapp-go-config.yaml"
    )
    
    local missing_files=()
    
    for file in "${required_files[@]}"; do
        if [[ ! -f "$file" ]]; then
            missing_files+=("$file")
        fi
    done
    
    if [[ ${#missing_files[@]} -eq 0 ]]; then
        log_success "All required documentation files are present"
        return 0
    else
        log_error "Missing required documentation files:"
        for file in "${missing_files[@]}"; do
            log_error "  - $file"
        done
        return 1
    fi
}

# Function to generate documentation statistics
generate_stats() {
    log_info "Generating documentation statistics"
    
    local stats_file="$DOCS_DIR/deployment/stats.txt"
    
    cat > "$stats_file" << EOF
Documentation Statistics - Generated on $TIMESTAMP

File Counts:
- Markdown files: $(find "$DOCS_DIR" -name "*.md" | wc -l)
- JSON files: $(find "$DOCS_DIR" -name "*.json" | wc -l)
- YAML files: $(find "$DOCS_DIR" -name "*.yaml" -o -name "*.yml" | wc -l)
- Text files: $(find "$DOCS_DIR" -name "*.txt" | wc -l)

File Sizes:
- Total documentation size: $(du -sh "$DOCS_DIR" | cut -f1)
- Complete Guide size: $(du -sh "$DOCS_DIR/COMPLETE_GUIDE.md" 2>/dev/null | cut -f1 || echo "N/A")
- LLM docs size: $(du -sh "$DOCS_DIR/llm/" 2>/dev/null | cut -f1 || echo "N/A")

Line Counts:
- Complete Guide: $(wc -l < "$DOCS_DIR/COMPLETE_GUIDE.md" 2>/dev/null || echo "N/A")
- LLM text format: $(wc -l < "$DOCS_DIR/llm/whatsapp-go-docs.txt" 2>/dev/null || echo "N/A")

Last Updated: $TIMESTAMP
EOF
    
    log_success "Documentation statistics saved to: $stats_file"
}

# Main execution function
main() {
    log_info "Starting documentation update process"
    log_info "Project root: $PROJECT_ROOT"
    log_info "Documentation directory: $DOCS_DIR"
    
    # Change to project root
    cd "$PROJECT_ROOT"
    
    # Check completeness
    if ! check_completeness; then
        log_error "Documentation completeness check failed"
        exit 1
    fi
    
    # Validate all documentation files
    log_info "Validating documentation files"
    
    # Validate markdown files
    while IFS= read -r -d '' file; do
        validate_markdown "$file" || exit 1
    done < <(find "$DOCS_DIR" -name "*.md" -print0)
    
    # Validate JSON files
    while IFS= read -r -d '' file; do
        validate_json "$file" || exit 1
    done < <(find "$DOCS_DIR" -name "*.json" -print0)
    
    # Validate YAML files
    while IFS= read -r -d '' file; do
        validate_yaml "$file" || exit 1
    done < <(find "$DOCS_DIR" -name "*.yaml" -o -name "*.yml" -print0)
    
    # Update timestamps
    update_timestamps
    
    # Generate statistics
    generate_stats
    
    log_success "Documentation update process completed successfully"
    log_info "Documentation is ready for deployment"
}

# Script usage information
usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -v, --validate Only validate files without updating"
    echo "  -s, --stats    Only generate statistics"
    echo ""
    echo "This script updates and validates the WhatsApp Go library documentation."
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        usage
        exit 0
        ;;
    -v|--validate)
        log_info "Running validation only"
        check_completeness
        exit $?
        ;;
    -s|--stats)
        log_info "Generating statistics only"
        generate_stats
        exit 0
        ;;
    "")
        main
        ;;
    *)
        log_error "Unknown option: $1"
        usage
        exit 1
        ;;
esac
