// +build linux

/*
 * struct_linux.go
 *
 * Copyright 2017 Bill Zissimopoulos
 */
/*
 * This file is part of Cgofuse.
 *
 * It is licensed under the MIT license. The full license text can be found
 * in the License.txt file at the root of this project.
 */

package fs

func getMode(isDir bool) uint32 {
	if isDir {
		return 16895
	}
	return 33279
}