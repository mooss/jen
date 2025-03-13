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
GAP_HEIGHT = .4;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

// Width of the frame relative to the square.
FRAME_RATIO = .8;

// Radius of the inner circles.
INNER_CIRCLE_RADIUS = 7;

// Scale factor for the base to be able to clip the parts together.
FIT_RATIO = .995;

// Size of the squares on the corner that are used to cut the corners of the base.
CHAMFER_SQUARE_SIZE = 11;

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

// Create a single square for corners using global parameters.
module corner_square(x, y) {
	translate([x, y, 0])
		rotate([0, 0, 45])
		square(CHAMFER_SQUARE_SIZE, center=true);
}

// Place squares at the four corners of the base.
module corner_squares() {
	hsq = SQUARE_SIZE/2;

	corner_square(-hsq,  hsq); // Top left.
	corner_square( hsq,  hsq); // Top right.
	corner_square(-hsq, -hsq); // Bottom left.
	corner_square( hsq, -hsq); // Bottom right.
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

module inner_chamfers() {
	intersection() {
		rounded_square(SQUARE_SIZE, CORNER_RADIUS);
		corner_squares();
	}
}

// Assemble the frame and the inner chamfers.
module chamfered_frame() {
	frame();
	inner_chamfers();
}

// Bottom part and plateau of the token.
module base(xy, base_height) {
	linear_extrude(height=base_height+PLATEAU_HEIGHT)
		difference(){
		rounded_square(PLATEAU_SIZE, CORNER_RADIUS);
		inner_chamfers();
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

scale([FIT_RATIO, FIT_RATIO, 1])
base(SQUARE_SIZE, BASE_HEIGHT);

translate([1.2*SQUARE_SIZE, 0, 0])
assembled_frame();
