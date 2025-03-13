////////////////
// Parameters //

// Size of the square used as the foundation of the token.
SQUARE_SIZE = 25;

// Size of the plateau, the central part of the base that is higher than the border.
PLATEAU_SIZE = 22;

// Height of the part of the frame that is above the plateau.
FRAME_HEIGHT = .6;

// Total height of the token.
TOTAL_HEIGHT = 2.5;

// Height of the part of the plateau that is above the base.
PLATEAU_HEIGHT = TOTAL_HEIGHT - FRAME_HEIGHT;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

// Width of the frame relative to the square.
FRAME_RATIO = .8;

// Radius of the inner circles.
INNER_CIRCLE_RADIUS = 7;

// Scale factor for the base to be able to clip the parts together.
FIT_RATIO = .997;

///////////////////////
// Shape definitions //

// Make a square with rounded circles.
module rounded_square(xy, corner_radius) {
	minkowski() {
		square(xy - 2 * corner_radius, center = true);
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

// Create a single circle for corners using global parameters.
module corner_circle(x, y) {
	translate([x, y, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);
}

// Place circles at the four corners of the base.
module corner_circles() {
	hsq = SQUARE_SIZE/2;

	corner_circle(-hsq,  hsq); // Top left.
	corner_circle( hsq,  hsq); // Top right.
	corner_circle(-hsq, -hsq); // Bottom left.
	corner_circle( hsq, -hsq); // Bottom right.
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

// Bottom part of the token.
// This part is rounded to minimize adhesion problems when printing it.
module base(xy) {
	linear_extrude(height=PLATEAU_HEIGHT)
		rounded_square(PLATEAU_SIZE, CORNER_RADIUS);
}

// Frame with a middle rectangular part remove to fit the base inside it.
// The removed part is rectangular and not the rounded base to make it as easy as possible to fit a
// piece of paper.
module assembled_frame() {
	difference() {
		linear_extrude(height=TOTAL_HEIGHT)
			circled_frame();
		translate([0, 0, -.5])
			linear_extrude(height=PLATEAU_HEIGHT+.5)
			square(PLATEAU_SIZE, center=true);
	}
}


////////////////////
// Shape assembly //

// rounded_square(SQUARE_SIZE, CORNER_RADIUS);
// frame();
// corner_circles();
// inner_circles();
// circled_frame();

scale(FIT_RATIO)
base(SQUARE_SIZE);

translate([1.2*SQUARE_SIZE, 0, 0])
assembled_frame();
