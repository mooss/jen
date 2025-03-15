////////////////
// Parameters //

// Size of the rectangle used as the foundation of the token.
TOKEN = [21, 30];

// Total height of the token.
TOKEN_HEIGHT = 2.5;

// Height of the core.
CORE_HEIGHT = 1.9;

// Radius of the rounded corners.
CORNER_RADIUS = 1;

// Number of segments for all the circles (rounded corners, inner circles).
$fn = 64;

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
		circle(r = corner_radius);
	}
}

// The maximal X/Y footprint of the token.
module flat_token(center=true) {
	rounded_rect(TOKEN.x, TOKEN.y, CORNER_RADIUS, center=center);
}

function addv2s(v, s) = [v[0]+s, v[1]+s];

////////////////
// Token core //
// The bottom part of the token, which presses the paper against the frame.

// Scale factor for the core, allowing the parts to fit together.
CORE_FIT_RATIO = .999;

// Distance between the core and each side of the token.
CORE_PADDING = 1;

// Size of the core.
CORE = [TOKEN.x - 2*CORE_PADDING, TOKEN.y - 2*CORE_PADDING];

// Bottom part of the token.
// This part is rounded to minimize adhesion problems when printing it.
module core() {
	linear_extrude(height=CORE_HEIGHT)
		rounded_rect(CORE.x, CORE.y, CORNER_RADIUS);
}

/////////////////
// Token frame //
// The top and side part of the token, with a central void that exposes the paper underneath.

// Width of the frame's border (the part against which the paper is pressed by the core).
// paper is pressed).
FRAME_BORDER = .75;

// Size of the empty part of the frame.
FRAME_VOID = [CORE.x - 2*FRAME_BORDER, CORE.y - 2*FRAME_BORDER];

// Subtract a smaller version of the rounded square from itself to create the fundamental block of a
// rounded frame.
module frame_block() {
	difference() {
		flat_token();
		rounded_rect(FRAME_VOID.x, FRAME_VOID.y, CORNER_RADIUS * (FRAME_VOID.x/TOKEN.x + FRAME_VOID.y/TOKEN.y)/2);
	}
}

// Radius of the inner circles (the circles that extend inwards from each corner at the top of the
// frame).
INNER_CIRCLE_RADIUS = 4; // TODO: think about removing this.

// Create a single circle for corners.
module corner_circle(x, y) {
	translate([x, y, 0])
		circle(r = INNER_CIRCLE_RADIUS);
}

// Place circles at the four corners of the frame.
module corner_circles() {
	x = TOKEN.x/2;
	y = TOKEN.y/2;

	corner_circle(-x,  y); // Top left.
	corner_circle( x,  y); // Top right.
	corner_circle(-x, -y); // Bottom left.
	corner_circle( x, -y); // Bottom right.
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
			rectangle(CORE.x, CORE.y);
	}
}

//////////
// Grid //
// The 2d grid on which the tokens can be placed.

// Number of cells.
GRID = [4, 4];

// Distance between each cell.
INTER_CELL = 2;

// Height of the grid at the lowest point, where the token can be inserted.
GRID_LOW_HEIGHT = 1;

// Height of the grid at the highest point.
GRID_HEIGHT = GRID_LOW_HEIGHT + TOKEN_HEIGHT;

// Space between neighboring cells.
CELL_DIST = addv2s(TOKEN, INTER_CELL);

// Dimensions of the grid.
GRID_DIM = addv2s([CELL_DIST.x * GRID.x, CELL_DIST.y * GRID.y], INTER_CELL);

// Cell size increase to make the tokens fit into the cells.
CELL_FIT_RATIO = [1.001, 1.001, 1.15];

// Size of the empty space beneath each cell.
VOID = addv2s(TOKEN, -7);

// Fundamental block of the grid from which cells can be substracted.
module grid_block() {
	linear_extrude(height=GRID_HEIGHT)
		rounded_rect(GRID_DIM.x, GRID_DIM.y, CORNER_RADIUS, center=false);
}

// Repeats its children on all the cells.
module foreach_cell() {
	for(x = [0:GRID.x-1])
		for(y = [0:GRID.y-1]) {
			translate([x*CELL_DIST.x + INTER_CELL, y*CELL_DIST.y + INTER_CELL])
				children();
		}
}

// Matrix of all the cells to carve from the grid block.
module grid_cells() {
	foreach_cell()
		linear_extrude(height=TOKEN_HEIGHT)
		flat_token(center=false);
}

// Void placed in the middle of each cell.
module void_cells() {
	center = (TOKEN-VOID) / 2;
	foreach_cell()
		translate([center[0], center[1], -.1])
		linear_extrude(height=GRID_LOW_HEIGHT+.2)
		rounded_rect(VOID.x, VOID.y, CORNER_RADIUS, center=false);
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
		void_cells();
	}
}

////////////////////
// Shape assembly //

scale(CORE_FIT_RATIO)
core();

translate([1.2*TOKEN.x, 0, 0])
assembled_frame();

translate([0, TOKEN.y, 0])
assembled_grid();
