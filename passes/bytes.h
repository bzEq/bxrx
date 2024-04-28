// Copyright (c) 2023 Kai Luo <gluokai@gmail.com>. All rights reserved.

#ifndef BX_BYTES_H
#define BX_BYTES_H

#include <stddef.h>

void ByteSwap(void *__restrict__ dst, const void *__restrict__ src, size_t len);
void ByteSwapInPlace(void *restrict, size_t);

#endif
