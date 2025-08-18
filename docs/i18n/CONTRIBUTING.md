# i18n Contributing Guide

This guide explains how to add new language support or improve existing translations in GOTRS.

## Table of Contents
- [Quick Start](#quick-start)
- [Translation Structure](#translation-structure)
- [Adding a New Language](#adding-a-new-language)
- [Testing Translations](#testing-translations)
- [API Endpoints](#api-endpoints)
- [Best Practices](#best-practices)

## Quick Start

To add or improve translations:

1. Navigate to `/internal/i18n/translations/`
2. Edit existing language files or create new ones
3. Test your translations using the validation tools
4. Submit a PR with 100% coverage

## Translation Structure

Translation files are JSON files with nested structure:

```json
{
  "app": {
    "name": "GOTRS",
    "title": "GOTRS - Ticketing System"
  },
  "navigation": {
    "dashboard": "Dashboard",
    "tickets": "Tickets"
  }
}
```

### Key Naming Convention
- Use lowercase with underscores: `ticket_created`
- Group related keys: `tickets.new_ticket`
- Use consistent prefixes: `button.save`, `label.email`

## Adding a New Language

### Step 1: Create Translation File

1. Copy `en.json` as template:
```bash
cp internal/i18n/translations/en.json internal/i18n/translations/xx.json
```

2. Replace `xx` with your language code (ISO 639-1)

### Step 2: Translate All Keys

Use the API to check missing keys:
```bash
curl http://localhost:8080/api/v1/i18n/missing/xx
```

### Step 3: Validate Completeness

Run validation to ensure 100% coverage:
```bash
# Use gotrs-babelfish for validation
make babelfish-validate LANG=xx

# Or run directly
docker exec gotrs-backend go run cmd/gotrs-babelfish/main.go -action=validate -lang=xx
```

Or use the API:
```bash
curl http://localhost:8080/api/v1/i18n/validate/xx
```

### Step 4: Test in Application

Test your translations by adding `?lang=xx` to any URL:
```
http://localhost:8080/dashboard?lang=xx
```

## Testing Translations

### Using gotrs-babelfish (The Universal Translation Tool)

```bash
# Run comprehensive translation tests
docker exec gotrs-backend go test ./internal/i18n -v

# Use gotrs-babelfish for coverage analysis
make babelfish-coverage

# Or run directly with custom options
docker exec gotrs-backend go run cmd/gotrs-babelfish/main.go -action=coverage
```

### Coverage Requirements

- **100% coverage required** - All English keys must have translations
- **No extra keys** - Don't add keys not present in English
- **Format consistency** - Maintain placeholder formatting (%s, %d)

## API Endpoints

GOTRS provides comprehensive i18n management APIs:

### Get Translation Coverage
```bash
GET /api/v1/i18n/coverage
```
Returns coverage statistics for all languages.

### Get Missing Keys
```bash
GET /api/v1/i18n/missing/{lang}
```
Lists all missing translation keys for a language.

### Export Translations
```bash
GET /api/v1/i18n/export/{lang}?format=json
GET /api/v1/i18n/export/{lang}?format=csv
```
Export translations in JSON or CSV format.

### Validate Translations
```bash
GET /api/v1/i18n/validate/{lang}
```
Validates translation completeness and correctness.

### Example: Check German Coverage
```bash
curl http://localhost:8080/api/v1/i18n/coverage | jq '.languages[] | select(.code=="de")'
```

## Best Practices

### 1. Context Matters
Consider where text appears when translating:
- Button text should be concise
- Help text can be more descriptive
- Error messages should be clear and actionable

### 2. Maintain Consistency
- Use consistent terminology throughout
- Follow the glossary for technical terms
- Keep formatting consistent with English

### 3. Handle Placeholders
Preserve placeholders in translations:
```json
"min_length": "Minimum length is %d characters"
```

### 4. Cultural Adaptation
- Use appropriate date/time formats
- Consider text direction (RTL languages)
- Adapt idioms and examples

### 5. Testing Process
1. Complete all translations
2. Run validation tests
3. Test in application UI
4. Review with native speakers

## Translation Guidelines

### General Rules
- Keep translations natural and fluent
- Don't translate literally if it sounds unnatural
- Maintain professional tone
- Use formal/informal address consistently

### Technical Terms
Some terms should remain in English:
- API
- URL
- Email
- Admin
- Dashboard (optional)

### Common Patterns

#### Status Messages
```json
"status": {
  "new": "New",        // Keep concise
  "open": "Open",      // Single word preferred
  "closed": "Closed"   // Past participle for completed states
}
```

#### Form Labels
```json
"labels": {
  "email": "Email Address",     // Be specific
  "phone": "Phone Number",      // Include "Number" for clarity
  "required": "Required Field"  // Clear indicators
}
```

#### Error Messages
```json
"errors": {
  "not_found": "Resource not found",              // What happened
  "unauthorized": "You are not authorized",       // Why it happened
  "try_again": "Please try again later"          // What to do
}
```

## Contributing Process

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b i18n/add-xx-language`
3. **Add translations**: Follow the structure of `en.json`
4. **Test thoroughly**: Use validation tools and UI testing
5. **Submit PR**: Include coverage report in description

### PR Checklist
- [ ] 100% translation coverage
- [ ] No extra keys beyond English
- [ ] Validation tests pass
- [ ] UI tested with new language
- [ ] Native speaker review (preferred)

## Tools and Resources

### Development Tools
- **VS Code Extensions**: JSON language support, i18n Ally
- **Online Tools**: Google Translate (for reference only)
- **Validation**: Built-in test suite

### Command Reference
```bash
# Test all languages with gotrs-babelfish
make babelfish-coverage

# Check specific language
make babelfish-validate LANG=de

# Export for translation service
docker exec gotrs-backend go run cmd/gotrs-babelfish/main.go \
  -action=export -lang=en -file=/tmp/en.csv -format=csv

# Import translations
docker exec gotrs-backend go run cmd/gotrs-babelfish/main.go \
  -action=import -lang=fr -file=/tmp/fr.csv -format=csv

# API alternatives
curl http://localhost:8080/api/v1/i18n/coverage
curl http://localhost:8080/api/v1/i18n/export/en?format=csv > en.csv
curl http://localhost:8080/api/v1/i18n/validate/xx
```

## Getting Help

- **Discord**: Join our community for translation discussions
- **GitHub Issues**: Report translation bugs or suggestions
- **API Documentation**: See `/api/docs` for full API reference

## Language Status

Current language support and coverage:

| Language | Code | Coverage | Status |
|----------|------|----------|--------|
| English | en | 100% | ✅ Complete |
| German | de | 100% | ✅ Complete |
| Spanish | es | ~44% | 🚧 In Progress |
| French | fr | ~44% | 🚧 In Progress |
| Portuguese | pt | ~44% | 🚧 In Progress |
| Japanese | ja | ~44% | 🚧 In Progress |
| Chinese | zh | ~44% | 🚧 In Progress |

Help us reach 100% coverage for all languages!