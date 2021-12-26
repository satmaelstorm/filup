package domain

import "strconv"

func init() {

}

func ChunkFileName(uid string, num int) string {
	return uid + "_part_" + strconv.Itoa(num)
}

func MetaFileName(uid string) string {
	return uid + "_meta"
}
