# Cache-Aside Testing Service

This repository is part of the **main service**, available at [Sidokaredev Repository](https://github.com/Sidokaredev/future-interns-backend). It is specifically designed to test the **cache-aside** pattern using **Redis**.

## Overview

This service implements the **cache-aside** pattern, where data is first checked in Redis before querying the database. The goal is to evaluate cache performance and its impact on response times.

## Features

- Implements **cache-aside** pattern with Redis  
- Measures cache hit/miss rates  
- Response times
- Resource utilization (CPU and Memory usage pre-request)
- Uses **GORM** for database interactions (if applicable)