////////////////
// Parameters //

// Size of the square used as the foundation of the token.
TOKEN_SIZE = 25;

// Size of the core.
CORE_SIZE = 22;

// Total height of the token.
TOKEN_HEIGHT = 2.5;

// Height of the core.
CORE_HEIGHT = 1.9;

// Radius of the rounded corners.
CORNER_RADIUS = 3;

// Number of segments for all the circles (rounded corners, inner circles).
CIRCLE_RESOLUTION = 128;

////////////////
// Primitives //
// Shapes that are not specific to this project, but generally useful to implement it.

// A regular rectangle.
module rectangle(x, y, center=true) {
	scale([x, y, 1])
		square(1, center=center);
	// square(1, center=true);
}

// A rectangle with rounded corners.
module rounded_rect(x, y, corner_radius, center=true) {
	minkowski() {
		// To "compensate" for the minkowski sum, xy must be adjusted.
		rectangle(x - 2 * corner_radius, y - 2 * corner_radius, center=center);
		circle(r = corner_radius, $fn = CIRCLE_RESOLUTION);
	}
}

// A square with rounded corners.
module rounded_square(xy, corner_radius, center=true) {
	rounded_rect(xy, xy, corner_radius, center=center);
}

// The maximal X/Y footprint of the token.
module flat_token(center=true) {
	rounded_square(TOKEN_SIZE, CORNER_RADIUS, center=center);
}

////////////////
// Token core //
// The bottom part of the token, which presses the paper against the frame.

// Scale factor for the core, allowing the parts to fit together.
CORE_FIT_RATIO = .997;

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
		flat_token();
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
		flat_token();
		corner_circles();
	}
}

// Frame with a middle rectangular part removed to fit the core inside it.
// The removed part is rectangular and not rounded to make it as easy as possible to fit a piece of
// paper.
module assembled_frame() {
	difference() {
		linear_extrude(height=TOKEN_HEIGHT) {
			frame_block();
			inner_circles();
		}
		translate([0, 0, -.5])
			linear_extrude(height=CORE_HEIGHT+.5)
			square(CORE_SIZE, center=true);
	}
}

//////////
// Grid //
// The 2d grid on which the tokens can be placed.

// Number of cells.
GRID_X = 1; GRID_Y = 2;

// Distance between each cell.
INTER_CELL = 2;

// Height of the grid at the lowest point, where the token can be inserted.
GRID_LOW_HEIGHT = 2;

// Height of the grid at the highest point.
GRID_HEIGHT = GRID_LOW_HEIGHT + TOKEN_HEIGHT;

// Space between neighboring cells.
CELL_DISTANCE = (INTER_CELL + TOKEN_SIZE);

// Dimensions of the grid.
GRID_DIM_X = CELL_DISTANCE * GRID_X + INTER_CELL;
GRID_DIM_Y = CELL_DISTANCE * GRID_Y + INTER_CELL;

// Cell size increase to make the tokens fit into the cells.
CELL_FIT_RATIO = 1.005;

// Fundamental block of the grid from which cells can be substracted.
module grid_block() {
	linear_extrude(height=GRID_HEIGHT)
		rounded_rect(GRID_DIM_X, GRID_DIM_Y, CORNER_RADIUS, center=false);
}

// Matrix of all the cells to carve from the grid block.
module grid_cells() {
	for(x = [0:GRID_X-1])
		for(y = [0:GRID_Y-1]) {
			// Place each cell at its X/Y coordinates and elevate it to the top of the grid.
			translate([x*CELL_DISTANCE + INTER_CELL, y*CELL_DISTANCE + INTER_CELL])
				linear_extrude(height=TOKEN_HEIGHT)
				flat_token(center=false);
		}
}

// Grid block from which the cells have been subtracted.
module assembled_grid() {
	// Slightly distorts the grid dimensions, but it is the simplest way to apply the cell fit
	// multiplier and it changes nothing in practice.
	scale(CELL_FIT_RATIO)
		difference() {
		grid_block();
		translate([0, 0, GRID_LOW_HEIGHT+0.001])
			grid_cells();
	}
}

////////////////////
// Shape assembly //

scale(CORE_FIT_RATIO)
core(TOKEN_SIZE);

translate([1.2*TOKEN_SIZE, 0, 0])
assembled_frame();

translate([0, TOKEN_SIZE, 0])
assembled_grid();
