// Side of the square used as the foundation of the token.
SQUARE_SIZE = 20;

// Height of the token.
EXTRUSION_HEIGHT = 2;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

// Width of the frame relative to the square.
FRAME_RATIO=.9;

// Radius of the inner circles.
INNER_CIRCLE_RADIUS=4;

// Make a square with rounded circles.
module rounded_square(xy, radius) {
	minkowski() {
		square([xy - 2 * radius, xy - 2 * radius], center = true);
		circle(r = radius, $fn = CIRCLE_RESOLUTION);
	}
}

// Subtract a smaller version of the rounded square from itself to create a rounded frame.
module frame(xy, radius, cut_ratio) {
	difference() {
		rounded_square(xy, radius);

		// The cut ratio for the radius is prettier when put to the power of 3.
		// Why? IDK.
		rounded_square(xy * cut_ratio, radius * cut_ratio ^3);
	}
}

// Place circles at the four corners of the square.
module corner_circles(xy, radius) {
	// Half XY.
	hxy = xy/2;

	// Top-left.
	translate([-hxy, hxy, 0])
		circle(r = radius, $fn = CIRCLE_RESOLUTION);

	// Top-right.
	translate([hxy, hxy, 0])
		circle(r = radius, $fn = CIRCLE_RESOLUTION);

	// Bottom-left.
	translate([-hxy, -hxy, 0])
		circle(r = radius, $fn = CIRCLE_RESOLUTION);

	// Bottom-right.
	translate([hxy, -hxy, 0])
		circle(r = radius, $fn = CIRCLE_RESOLUTION); 
}

// Keep the part of the circles that are inside the rounded square.
module inner_circles(xy, radius, cut_ratio) {
	intersection() {
		rounded_square(xy, cut_ratio);
		corner_circles(xy, radius);
	}
}

// Assemble the frame and the inner circles.
module circled_frame(xy, frame_radius, cut_ratio, circle_radius) {
	frame(xy, frame_radius, cut_ratio);
	inner_circles(xy, circle_radius, cut_ratio);
}

linear_extrude(height=EXTRUSION_HEIGHT) {
	// rounded_square(SQUARE_SIZE, CORNER_RADIUS);
	// frame(SQUARE_SIZE, CORNER_RADIUS, FRAME_RATIO);
	// corner_circles(SQUARE_SIZE, INNER_CIRCLE_RADIUS);
	// inner_circles(SQUARE_SIZE, INNER_CIRCLE_RADIUS, FRAME_RATIO);
	circled_frame(SQUARE_SIZE, CORNER_RADIUS, .9, INNER_CIRCLE_RADIUS);
}
