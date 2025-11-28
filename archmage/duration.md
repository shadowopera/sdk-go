# RFC: Duration Sharding Encoding Specification

## Abstract

This document specifies a compact encoding scheme for Go's `time.Duration` values using integer arrays to represent durations in their most efficient form based on precision requirements.

## 1. Format Specification

### 1.1 Zero Duration
- **Input**: Zero duration
- **Output**: `null`

### 1.2 Two-Element Array Format: `[type, value]`

#### 1.2.1 Second Precision (Type 0)
- **Condition**: Duration evenly divisible by 1 second (1e9 nanoseconds)
- **Format**: `[2]int64{0, seconds}`
- **Example**: 5 seconds → `[0, 5]`

#### 1.2.2 Millisecond Precision (Type 1)
- **Condition**: Duration evenly divisible by 1 millisecond (1e6 nanoseconds)
- **Format**: `[2]int64{1, milliseconds}`
- **Example**: 1500ms → `[1, 1500]`

#### 1.2.3 Microsecond Precision (Type 2)
- **Condition**: Duration evenly divisible by 1 microsecond (1e3 nanoseconds)
- **Format**: `[2]int64{2, microseconds}`
- **Example**: 2500μs → `[2, 2500]`

#### 1.2.4 Nanosecond Precision (Type 3)
- **Condition**: Duration less than 1 second
- **Format**: `[2]int64{3, nanoseconds}`
- **Example**: 6500ns → `[3, 6500]`

### 1.3 Three-Element Array Format: `[4, seconds, nanoseconds]`

#### 1.3.1 Mixed Precision (Type 4)
- **Condition**: Duration requires both second and sub-second components
- **Format**: `[3]int64{4, whole_seconds, remaining_nanoseconds}`
- **Example**: 1.00000002 seconds → `[4, 1, 200]`

## 2. Examples

```
Input: 0s              → Output: null
Input: 5s              → Output: [0, 5]
Input: 1500ms          → Output: [1, 1500]
Input: 2500μs          → Output: [2, 2500]
Input: 6500ns          → Output: [3, 6500]
Input: 1.0000002s      → Output: [4, 1, 200]
Input: 2m30.000070123s → Output: [4, 150, 70123]
```

This encoding optimizes storage by automatically selecting the most appropriate time unit while preserving complete precision information and providing superior cross-platform implementation ease compared to string-based duration formats.
