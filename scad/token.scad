////////////////
// Parameters //

// Size of the square used as the foundation of the token.
SQUARE_SIZE = 25;

// Size of the plateau, the central part of the base that is higher than the border.
PLATEAU_SIZE = 22;

// Height of the bottom part of the token.
BASE_HEIGHT = 0;

// Height of the part of the frame that is above the plateau.
FRAME_HEIGHT = .6;

// Total height of the token.
TOTAL_HEIGHT = 2.5;

// Height of the part of the plateau that is above the base.
PLATEAU_HEIGHT = TOTAL_HEIGHT - FRAME_HEIGHT - BASE_HEIGHT;

// Height of the paper gap, the place below the frame ledge where paper can be inserted.
GAP_HEIGHT = .2;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

// Width of the frame relative to the square.
FRAME_RATIO = .8;

// Radius of the inner circles.
INNER_CIRCLE_RADIUS = 6;

///////////////////////
// Shape definitions //

// Make a square with rounded circles.
module rounded_square(xy, corner_radius) {
	minkowski() {
		square([xy - 2 * corner_radius, xy - 2 * corner_radius], center = true);
		circle(r = corner_radius, $fn = CIRCLE_RESOLUTION);
	}
}

// Subtract a smaller version of the rounded square from itself to create a rounded frame.
module frame() {
	difference() {
		rounded_square(SQUARE_SIZE, CORNER_RADIUS);
		rounded_square(SQUARE_SIZE * FRAME_RATIO, CORNER_RADIUS * FRAME_RATIO);
	}
}

// Place circles at the four corners of the square.
module corner_circles() {
	// Half square size.
	hsq = SQUARE_SIZE/2;

	// Top-left.
	translate([-hsq, hsq, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);

	// Top-right.
	translate([hsq, hsq, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);

	// Bottom-left.
	translate([-hsq, -hsq, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);

	// Bottom-right.
	translate([hsq, -hsq, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);
}

// Keep the part of the circles that are inside the rounded square.
module inner_circles() {
	intersection() {
		rounded_square(SQUARE_SIZE, CORNER_RADIUS);
		corner_circles();
	}
}

// Assemble the frame and the inner circles.
module circled_frame() {
	frame();
	inner_circles();
}

// Bottom part and plateau of the token.
module base(xy, base_height) {
	linear_extrude(height=base_height+PLATEAU_HEIGHT)
		difference(){
		rounded_square(PLATEAU_SIZE, CORNER_RADIUS);
		inner_circles();
	}
}

module assembled_frame() {
	difference() {
		linear_extrude(height=TOTAL_HEIGHT)
			circled_frame();
		translate([0, 0, -.5])
			base(SQUARE_SIZE+.05, BASE_HEIGHT+.5);
		translate([0, 0, PLATEAU_HEIGHT-GAP_HEIGHT])
			linear_extrude(height=GAP_HEIGHT)
			rounded_square(PLATEAU_SIZE, CORNER_RADIUS);
	}
}


////////////////////
// Shape assembly //

// rounded_square(SQUARE_SIZE, CORNER_RADIUS);
// frame();
// corner_circles();
// inner_circles();
// circled_frame();

base(SQUARE_SIZE, BASE_HEIGHT);
translate([1.2*SQUARE_SIZE, 0, 0])
assembled_frame();
