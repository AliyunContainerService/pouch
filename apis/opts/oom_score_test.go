package opts

import "testing"

func Test_ValidateOOMScore_1(t *testing.T) {
	if error := ValidateOOMScore(500); error != nil { //try a unit test on function
		t.Error("ValidateOOMScore测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("第一个测试通过了") //记录一些你期望记录的信息
	}
}

func Test_ValidateOOMScore_2(t *testing.T) {
	if error := ValidateOOMScore(1000); error != nil { //try a unit test on function
		t.Error("ValidateOOMScore测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("第二个测试通过了") //记录一些你期望记录的信息
	}
}

func Test_ValidateOOMScore_3(t *testing.T) {
	if error := ValidateOOMScore(-1000); error != nil { //try a unit test on function
		t.Error("ValidateOOMScore测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("第三个测试通过了") //记录一些你期望记录的信息
	}
}

func Test_ValidateOOMScore_4(t *testing.T) {
	if error := ValidateOOMScore(1001); error == nil { //try a unit test on function
		t.Error("ValidateOOMScore测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("第四个测试通过了") //记录一些你期望记录的信息
	}
}

func Test_ValidateOOMScore_5(t *testing.T) {
	if error := ValidateOOMScore(-1001); error == nil { //try a unit test on function
		t.Error("ValidateOOMScore测试没通过") // 如果不是如预期的那么就报错
	} else {
		t.Log("第五个测试通过了") //记录一些你期望记录的信息
	}
}
