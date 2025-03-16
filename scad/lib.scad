// A regular rectangle.
module rectangle(x, y, center=true) {
	scale([x, y, 1])
		square(1, center=center);
}
