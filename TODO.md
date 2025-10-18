# Epic #6: Testing Strategy & Quality Gates - TODO

**Milestone:** v0.6 - Testing & Performance | **Status:** 0/14 completed

---

## Arbeitsreihenfolge (Bottom-Up Strategie)

### 🔴 Phase 1: Foundation (Start hier)

1. **#67** - Unit Test Coverage >80% (High) - `go test -cover ./...`
2. **#73** - Race Detector CI Integration (High) - `go test -race ./...`
3. **#74** - Test Data Fixtures (Medium) - `testdata/` Struktur + JSONL Samples
4. **#79** - Mock & Stub Utilities (Medium) - Mock DB/HTTP, Stub Logger

### 🟡 Phase 2: Component Integration

5. **#68** - Integration Tests (High) - Config+Logger, Parser+DB, Worker+Parser+DB
6. **#69** - E2E Tests (High) - Full Import Workflow, <30s Runtime

### 🟢 Phase 3: CI/CD & Automation

7. **#75** - CI/CD Pipeline (High) - Lint+Test+Coverage in GitHub Actions
8. **#70** - Performance Regression Tests (Medium) - Benchmark Baseline + CI
9. **#76** - Test Parallelization (Medium) - Package-level + `-parallel` Flag

### 🔵 Phase 4: Advanced (Optional)

10. **#71** - Fuzz Testing (Low) - go-fuzz, 100k iterations
11. **#72** - Property-Based Tests (Low) - gopter für Config/Parser/Batch
12. **#77** - Golden File Tests (Low) - Parser Output Verification
13. **#78** - Coverage Report & Badge (Low) - codecov.io Integration
14. **#80** - Testing Best Practices Guide (Low) - `docs/testing/README.md`

---

## Quick Start

```bash
# 1. Coverage Analyse
go test -cover ./...

# 2. Race Detector
go test -race ./...

# 3. Make Targets
make test
make lint
```

**GitHub Issues:** <https://github.com/Sternrassler/EVE-SDE-Database-Builder/milestone/6>
