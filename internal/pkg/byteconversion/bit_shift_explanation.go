// Copyright 2023 Hoster Authors. All rights reserved.
// Use of this source code is governed by an Apache License 2.0
// license that can be found in the LICENSE file.

package byteconversion

// The line multiplier = 1 << 40 is a bit-shifting operation in Go.

// Here's an explanation of what it does:
// The << operator is the left shift operator.
// In this case, it's used to shift the binary representation of the integer 1 to the left by 40 positions.

// In binary, the decimal number 1 is represented as 0000000000000000000000000000000000000000000000000000000000000001.
// When you shift it to the left by 40 positions, you get 0000001000000000000000000000000000000000000000000000000000000000.

// In terms of binary arithmetic, shifting to the left by 40 positions effectively multiplies the original number by 2^40 (2 raised to the power of 40).
// In other words, it calculates 1 * 2^40, which is equal to 1 terabyte (TB).

// So, multiplier = 1 << 40 sets the multiplier variable to the number of bytes in a terabyte.
// This is used in the ConvertSizeToBytes function to scale the size based on the provided suffixes (TB, GB, MB, KB).
