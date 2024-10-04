// amd64 - go 1.23.0

#include "textflag.h"

TEXT ·asm_std(SB), $0
	XORQ DX, DX                // Slice offset.
	XORQ R9, R9                // Index.
	XORQ R8, R8                // Did find.
	XORQ R11, R11 		// Mask

	MOVQ key+24(FP), AX        // Load `key` parameter into X1.
	MOVQ keys+0(FP), BX        // Load address of the slice into BX (keys.data).
	MOVQ keys_len+16(FP), CX   // Load the length of the slice into CX (keys.len).

loop:
	CMPQ R9, CX
	JE done

	MOVQ 0(BX)(DX*1), R11
	ADDQ $8, DX

	CMPQ AX, R11
	JE found

	INCQ R9
	JMP loop

found:
	MOVQ $1, R8       // Return 1 if found.

done:
	// Set the index where the match was found.
	MOVB R9, ret+32(FP)       // Return the index of the found key or 0 if not found.
	MOVQ R8, ret+33(FP)       // Return 1 if found.
	RET














// Register usage:
// R9: Result and temporary $0 register.
// SI: Pointer to the slice.
// CX: Length of the slice.
// X0: Key to find.
// Y0: Broadcasted key.
// R11: Temporary register for index when found.
// R12: Temporary register for mask when found.
// Y1: Temporary register for SIMD comparison.
// AX: Temporary register for SIMD comparison.
// DI: Loop index.
TEXT ·simd_find_idx(SB), $0-32
	// Load function parameters
	MOVQ key+24(FP), X0 		// Load key into X0.
	MOVQ keys+0(FP), SI 		// SI = &keys[0].
	MOVQ keys_len+16(FP), CX 	// CX = len(keys).
	SHLQ $3, CX 			// Multiply by 8 for convenience.

	VPBROADCASTQ X0, Y0 		// Broadcast key across YMM0.

	// Initialize variables
	XORQ R9, R9 			// R9 = 0 (return value).
	XORQ DI, DI              // DI = 0 (element index)

	XORQ R11, R11

loop_simd:
	// Magic hax:
	VMOVDQU64 (SI)(DI*1), Y1
	VPCMPEQQ  Y0, Y1, Y1      // Y1 = (Y1 == Y0)
	VPMOVMSKB Y1, AX         // Move mask to AX

	CMPQ    AX, R9
	CMOVQNE DI, R11
	CMOVQNE AX, R12

	ADDQ      $32, DI              // Move to next group of four elements
	CMPQ DI, CX
	JG   done
	JMP  loop_simd

done:
	TESTQ R11, R11
	JZ fail

success:
	SHRQ $3, R11             // Divide by 8 to get the index of the found element

	BSFQ R12, R9              // Calculate the index of the found element
	SHRQ $3, R9              // Divide by 8 to get the index of the found element

	ADDQ R11, R9              // Add the current index to the found index
	INCQ R9

	MOVQ    R9, ret+32(FP)      // Return index
	RET

fail:
	MOVQ    R9, ret+32(FP)      // Return index
	RET
