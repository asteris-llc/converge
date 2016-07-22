package skittles

import (
	"testing"
)

func TestBlack(t *testing.T) {
	in := "I'm a black string"
	out := "\033[30mI'm a black string\033[0m"
	if str := Black(in); str != out {
		t.Errorf("Black(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestRed(t *testing.T) {
	in := "I'm a red string"
	out := "\033[31mI'm a red string\033[0m"
	if str := Red(in); str != out {
		t.Errorf("Red(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestGreen(t *testing.T) {
	in := "I'm a green string"
	out := "\033[32mI'm a green string\033[0m"
	if str := Green(in); str != out {
		t.Errorf("Green(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestYellow(t *testing.T) {
	in := "I'm a yellow string"
	out := "\033[33mI'm a yellow string\033[0m"
	if str := Yellow(in); str != out {
		t.Errorf("Yellow(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlue(t *testing.T) {
	in := "I'm a blue string"
	out := "\033[34mI'm a blue string\033[0m"
	if str := Blue(in); str != out {
		t.Errorf("Blue(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestMagenta(t *testing.T) {
	in := "I'm a magenta string"
	out := "\033[35mI'm a magenta string\033[0m"
	if str := Magenta(in); str != out {
		t.Errorf("Magenta(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestCyan(t *testing.T) {
	in := "I'm a cyan string"
	out := "\033[36mI'm a cyan string\033[0m"
	if str := Cyan(in); str != out {
		t.Errorf("Cyan(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestWhite(t *testing.T) {
	in := "I'm a white string"
	out := "\033[37mI'm a white string\033[0m"
	if str := White(in); str != out {
		t.Errorf("White(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldBlack(t *testing.T) {
	in := "I'm a black string"
	out := "\033[1;30mI'm a black string\033[0m"
	if str := BoldBlack(in); str != out {
		t.Errorf("BoldBlack(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldRed(t *testing.T) {
	in := "I'm a red string"
	out := "\033[1;31mI'm a red string\033[0m"
	if str := BoldRed(in); str != out {
		t.Errorf("BoldRed(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldGreen(t *testing.T) {
	in := "I'm a green string"
	out := "\033[1;32mI'm a green string\033[0m"
	if str := BoldGreen(in); str != out {
		t.Errorf("BoldGreen(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldYellow(t *testing.T) {
	in := "I'm a yellow string"
	out := "\033[1;33mI'm a yellow string\033[0m"
	if str := BoldYellow(in); str != out {
		t.Errorf("BoldYellow(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldBlue(t *testing.T) {
	in := "I'm a blue string"
	out := "\033[1;34mI'm a blue string\033[0m"
	if str := BoldBlue(in); str != out {
		t.Errorf("BoldBlue(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldMagenta(t *testing.T) {
	in := "I'm a magenta string"
	out := "\033[1;35mI'm a magenta string\033[0m"
	if str := BoldMagenta(in); str != out {
		t.Errorf("BoldMagenta(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldCyan(t *testing.T) {
	in := "I'm a cyan string"
	out := "\033[1;36mI'm a cyan string\033[0m"
	if str := BoldCyan(in); str != out {
		t.Errorf("BoldCyan(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBoldWhite(t *testing.T) {
	in := "I'm a white string"
	out := "\033[1;37mI'm a white string\033[0m"
	if str := BoldWhite(in); str != out {
		t.Errorf("BoldWhite(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkBlack(t *testing.T) {
	in := "I'm a black string"
	out := "\033[5;30mI'm a black string\033[0m"
	if str := BlinkBlack(in); str != out {
		t.Errorf("BlinkBlack(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkRed(t *testing.T) {
	in := "I'm a red string"
	out := "\033[5;31mI'm a red string\033[0m"
	if str := BlinkRed(in); str != out {
		t.Errorf("BlinkRed(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkGreen(t *testing.T) {
	in := "I'm a green string"
	out := "\033[5;32mI'm a green string\033[0m"
	if str := BlinkGreen(in); str != out {
		t.Errorf("BlinkGreen(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkYellow(t *testing.T) {
	in := "I'm a yellow string"
	out := "\033[5;33mI'm a yellow string\033[0m"
	if str := BlinkYellow(in); str != out {
		t.Errorf("BlinkYellow(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkBlue(t *testing.T) {
	in := "I'm a blue string"
	out := "\033[5;34mI'm a blue string\033[0m"
	if str := BlinkBlue(in); str != out {
		t.Errorf("BlinkBlue(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkMagenta(t *testing.T) {
	in := "I'm a magenta string"
	out := "\033[5;35mI'm a magenta string\033[0m"
	if str := BlinkMagenta(in); str != out {
		t.Errorf("BlinkMagenta(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkCyan(t *testing.T) {
	in := "I'm a cyan string"
	out := "\033[5;36mI'm a cyan string\033[0m"
	if str := BlinkCyan(in); str != out {
		t.Errorf("BlinkCyan(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestBlinkWhite(t *testing.T) {
	in := "I'm a white string"
	out := "\033[5;37mI'm a white string\033[0m"
	if str := BlinkWhite(in); str != out {
		t.Errorf("BlinkWhite(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineBlack(t *testing.T) {
	in := "I'm a black string"
	out := "\033[4;30mI'm a black string\033[0m"
	if str := UnderlineBlack(in); str != out {
		t.Errorf("UnderlineBlack(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineRed(t *testing.T) {
	in := "I'm a red string"
	out := "\033[4;31mI'm a red string\033[0m"
	if str := UnderlineRed(in); str != out {
		t.Errorf("UnderlineRed(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineGreen(t *testing.T) {
	in := "I'm a green string"
	out := "\033[4;32mI'm a green string\033[0m"
	if str := UnderlineGreen(in); str != out {
		t.Errorf("UnderlineGreen(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineYellow(t *testing.T) {
	in := "I'm a yellow string"
	out := "\033[4;33mI'm a yellow string\033[0m"
	if str := UnderlineYellow(in); str != out {
		t.Errorf("UnderlineYellow(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineBlue(t *testing.T) {
	in := "I'm a blue string"
	out := "\033[4;34mI'm a blue string\033[0m"
	if str := UnderlineBlue(in); str != out {
		t.Errorf("UnderlineBlue(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineMagenta(t *testing.T) {
	in := "I'm a magenta string"
	out := "\033[4;35mI'm a magenta string\033[0m"
	if str := UnderlineMagenta(in); str != out {
		t.Errorf("UnderlineMagenta(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineCyan(t *testing.T) {
	in := "I'm a cyan string"
	out := "\033[4;36mI'm a cyan string\033[0m"
	if str := UnderlineCyan(in); str != out {
		t.Errorf("UnderlineCyan(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestUnderlineWhite(t *testing.T) {
	in := "I'm a white string"
	out := "\033[4;37mI'm a white string\033[0m"
	if str := UnderlineWhite(in); str != out {
		t.Errorf("UnderlineWhite(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseBlack(t *testing.T) {
	in := "I'm a black string"
	out := "\033[7;30mI'm a black string\033[0m"
	if str := InverseBlack(in); str != out {
		t.Errorf("InverseBlack(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseRed(t *testing.T) {
	in := "I'm a red string"
	out := "\033[7;31mI'm a red string\033[0m"
	if str := InverseRed(in); str != out {
		t.Errorf("InverseRed(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseGreen(t *testing.T) {
	in := "I'm a green string"
	out := "\033[7;32mI'm a green string\033[0m"
	if str := InverseGreen(in); str != out {
		t.Errorf("InverseGreen(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseYellow(t *testing.T) {
	in := "I'm a yellow string"
	out := "\033[7;33mI'm a yellow string\033[0m"
	if str := InverseYellow(in); str != out {
		t.Errorf("InverseYellow(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseBlue(t *testing.T) {
	in := "I'm a blue string"
	out := "\033[7;34mI'm a blue string\033[0m"
	if str := InverseBlue(in); str != out {
		t.Errorf("InverseeBlue(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseMagenta(t *testing.T) {
	in := "I'm a magenta string"
	out := "\033[7;35mI'm a magenta string\033[0m"
	if str := InverseMagenta(in); str != out {
		t.Errorf("InverseMagenta(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseCyan(t *testing.T) {
	in := "I'm a cyan string"
	out := "\033[7;36mI'm a cyan string\033[0m"
	if str := InverseCyan(in); str != out {
		t.Errorf("InverseCyan(%s) output %s. It should have been %s", in, str, out)
	}
}

func TestInverseWhite(t *testing.T) {
	in := "I'm a white string"
	out := "\033[7;37mI'm a white string\033[0m"
	if str := InverseWhite(in); str != out {
		t.Errorf("InverseWhite(%s) output %s. It should have been %s", in, str, out)
	}
}
