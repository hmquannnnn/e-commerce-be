# Database Seeds

This directory contains seed data for development and testing.

## Files

- `01_vietnam_locations.sql` - Vietnam administrative divisions (cities, districts, wards)

## Usage

### Run all seeds

```bash
make seed
```

### Run specific seed

```bash
make seed-locations
```

### Via Docker (manual)

```bash
docker exec -i ecommerce-postgres-user psql -U postgres -d user_db < seeds/01_vietnam_locations.sql
```

### Via psql (manual)

```bash
psql -h localhost -p 5433 -U postgres -d user_db < seeds/01_vietnam_locations.sql
```

## Data Sources

### Vietnam Locations

- **Source**: Sample data from common provinces/cities
- **Complete data**: https://github.com/kenzouno1/DiaGioiHanhChinhVN
- **Coverage**:
  - 63 provinces/cities (sample: ~50)
  - Districts (sample: ~23)
  - Wards (sample: ~30)

## Adding New Seeds

1. Create new file: `seeds/0X_description.sql`
2. Add seed command to Makefile
3. Update this README

## Notes

- Seeds are **optional** and only for development/testing
- Don't run seeds in production
- Seed files are numbered for execution order
- Each seed should be idempotent (safe to run multiple times)
