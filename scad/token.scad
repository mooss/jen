////////////////
// Parameters //

// Size of the square used as the foundation of the token.
TOKEN_SIZE = 25;

// Size of the core.
CORE_SIZE = 22;

// Total height of the token.
TOTAL_HEIGHT = 2.5;

// Height of the core.
CORE_HEIGHT = 1.9;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

////////////////
// Primitives //
// Shapes that are not specific to this project, but generally useful to implement it.

// A square with rounded corners.
module rounded_square(xy, corner_radius) {
	minkowski() {
		square(xy - 2 * corner_radius, center = true);
		circle(r = corner_radius, $fn = CIRCLE_RESOLUTION);
	}
}

////////////////
// Token core //
// The bottom part of the token, which presses the paper against the frame.

// Scale factor for the core, allowing the parts to fit together.
FIT_RATIO = .997;

// Bottom part of the token.
// This part is rounded to minimize adhesion problems when printing it.
module core(xy) {
	linear_extrude(height=CORE_HEIGHT)
		rounded_square(CORE_SIZE, CORNER_RADIUS);
}

/////////////////
// Token frame //
// The top and side part of the token, with a central void that exposes the paper underneath.

// Width of the frame's central void.
FRAME_RATIO = .8;

// Subtract a smaller version of the rounded square from itself to create the fundamental block of a
// rounded frame.
module frame_block() {
	difference() {
		rounded_square(TOKEN_SIZE, CORNER_RADIUS);
		rounded_square(TOKEN_SIZE * FRAME_RATIO, CORNER_RADIUS * FRAME_RATIO);
	}
}

// Radius of the inner circles (the circles that extend inwards from each corner at the top of the
// frame).
INNER_CIRCLE_RADIUS = 7;

// Create a single circle for corners.
module corner_circle(x, y) {
	translate([x, y, 0])
		circle(r = INNER_CIRCLE_RADIUS, $fn = CIRCLE_RESOLUTION);
}

// Place circles at the four corners of the frame.
module corner_circles() {
	hsq = TOKEN_SIZE/2;

	corner_circle(-hsq,  hsq); // Top left.
	corner_circle( hsq,  hsq); // Top right.
	corner_circle(-hsq, -hsq); // Bottom left.
	corner_circle( hsq, -hsq); // Bottom right.
}

// Keep the part of the circles that are inside the rounded square.
module inner_circles() {
	intersection() {
		rounded_square(TOKEN_SIZE, CORNER_RADIUS);
		corner_circles();
	}
}

// Frame with a middle rectangular part removed to fit the core inside it.
// The removed part is rectangular and not rounded to make it as easy as possible to fit a piece of
// paper.
module assembled_frame() {
	difference() {
		linear_extrude(height=TOTAL_HEIGHT) {
			frame_block();
			inner_circles();
		}
		translate([0, 0, -.5])
			linear_extrude(height=CORE_HEIGHT+.5)
			square(CORE_SIZE, center=true);
	}
}

////////////////////
// Shape assembly //

// rounded_square(TOKEN_SIZE, CORNER_RADIUS);
// frame_block();
// corner_circles();
// inner_circles();

scale(FIT_RATIO)
core(TOKEN_SIZE);

translate([1.2*TOKEN_SIZE, 0, 0])
assembled_frame();
