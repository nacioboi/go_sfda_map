
// Register usage:
// R9: Result and temporary $0 register.
// R15: Temporary register for $1.
// SI: Pointer to the slice.
// CX: Length of the slice.
// X0: Key to find.
// Y0: Broadcasted key.
// Y1: Temporary register for SIMD comparison.
// R11: Temporary register for index when found.
// R12: Temporary register for mask when found.
// R13: Temporary register for flag when found.
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
	XORQ R13, R13

	MOVQ $1, R15

loop_simd:
	// Magic hax:
	VMOVDQU64 (SI)(DI*1), Y1
	VPCMPEQQ  Y0, Y1, Y1      // Y1 = (Y1 == Y0)
	VPMOVMSKB Y1, AX         // Move mask to AX

	CMPQ    AX, R9
	CMOVQNE DI, R11
	CMOVQNE AX, R12
	CMOVQNE R15, R13

	ADDQ      $32, DI              // Move to next group of four elements
	CMPQ DI, CX
	JG   done
	JMP  loop_simd

done:
	TESTQ R13, R13
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
