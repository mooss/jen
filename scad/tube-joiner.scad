////////////////
// Parameters //
////////////////

// Size of the cube forming the core of the tube joiner.
SIZE = 30;

// How deep the tube goes inside a given face.
TUBE_DEPTH = 8;

// Witdh of the tube.
TUBE_WIDTH = 10.2;

// How much should the base cube be truncated.
TRUNCATION_RATIO = .75;

// Height of the support circle.
SUPPORT_HEIGHT = .4;

// Radius of the support circle.
SUPPORT_RADIUS = 15;

// Number of fragments.
$fn = 64;

// With of the beaming tube.
BEAM_TUBE_WIDTH = 10;

// Distance between the ground and the cornerstone within the foot.
FOOT_OFFSET = 5;

// Scale of the truncube to be subtracted from the feets to make the cornerstones fit within them.
FOOT_TRUNCUBE_VOID_SCALE = 1.03;

////////////////
// Primitives //
////////////////

use <lib.scad>

AXES = [[1, 0, 0],  // x
		[0, 1, 0],  // y
		[0, 0, 1]]; // z

module octahedron(size) {
	points=[[ 1,  0,  0],  // Right.
			[ 0,  1,  0],  // Forward.
			[ 0,  0,  1],  // Up.
			[-1,  0,  0],  // Left.
			[ 0, -1,  0],  // Backward.
			[ 0,  0, -1]]; // Down.

	faces=[[2, 1, 0], // Top faces.
		   [2, 3, 1],
		   [2, 4, 3],
		   [2, 0, 4],
		   [5, 0, 1], // Bottom faces.
		   [5, 1, 3],
		   [5, 3, 4],
		   [5, 4, 0]];

	polyhedron(points = points * size, faces = faces);
}

// Create a cube with truncated corners, guided by ratio.
// 0 is a regular untruncated cube, 1 is the cube fully truncated, i.e. an octahedron.
module truncube(cube_size, ratio) {
	octahedron_size=cube_size*(1.5-ratio);
	intersection() {
		cube(cube_size, true);
		octahedron(octahedron_size);
	}
}

/////////////////
// Cornerstone //
/////////////////

// One tube resting inside the top part of the cube.
module tube() {
	translate([0, 0, SIZE/2 - TUBE_DEPTH])
		cylinder(TUBE_DEPTH+.001, r=TUBE_WIDTH/2, $fn=128);
}

// All 6 rotations of the tube to cover the positive and negative position of each axis.
module tubes() {
	for(axis = [0:2]) {
		rotate(AXES[axis]*90)
			tube();
		rotate(AXES[axis]*90)
			mirror([0, 0, 1])
			tube();
	}
}

// A truncated cube with a hole on each face where a tube can be inserted.
module cornerstone() {
	difference() {
		truncube(SIZE, TRUNCATION_RATIO);
		tubes();
	}
}

// Additional surface below the model to help with bed adhesion.
module support() {
	difference() {
		translate([0, 0, -SIZE/2])
			linear_extrude(height=SUPPORT_HEIGHT)
			circle(SUPPORT_RADIUS);
		translate([0, 0, -0.001])
			truncube(SIZE, TRUNCATION_RATIO);
	}
}

/////////////////
// Corner beam //
/////////////////

CORNER_LEN = 100;
CORNER_HEIGHT = 15;
EXT_VOID = 80;
CORNER_EXTENSION = [16.5, 40];
CORNER_TUBE_SHIFT = 5;
CORNER_SMOOTHING = 2;
TRANSVERSE_TUBE_SHIFT = 1;

module half_sphere(r) {
	intersection() {
		translate([-r, -r])
			cube(r*2);
		sphere(r);
	}
}

////////////
// Single //

module bl_corner_piece() {
	a = [0, 0];
	b = [CORNER_LEN, 0];
	c = [0, CORNER_LEN];
	linear_extrude(height=CORNER_HEIGHT)
		difference() {
		union() {
			polygon([a, b, c]);
			translate([CORNER_LEN-CORNER_EXTENSION.x, 0])
				rectangle(CORNER_EXTENSION.x, CORNER_EXTENSION.y, center=false);
		}
		polygon([a, [EXT_VOID, 0], [0, EXT_VOID]]);
	}
}

module bl_corner_piece_smooth3d() {
	resize([CORNER_LEN, CORNER_LEN, CORNER_HEIGHT])
	translate([CORNER_SMOOTHING, CORNER_SMOOTHING, CORNER_SMOOTHING/2])
	minkowski() {
		bl_corner_piece();
		// Round everything but the top.
		rotate([0, 180, 0])
		half_sphere(r=CORNER_SMOOTHING);
	}
}

FRAME_TUBE_OFFSET = TUBE_WIDTH/2 + CORNER_TUBE_SHIFT;

module frame_tube_x() {
	translate([0, FRAME_TUBE_OFFSET, TUBE_WIDTH*.15])
	rotate([0, 90, 0])
	cylinder(h=CORNER_LEN+1, r=TUBE_WIDTH/2);
}

module frame_tube_y() {
	mirror([1, 0, 0])
	rotate([0, 0, 90])
		frame_tube_x();
}

module transverse_tube() {
	translate([CORNER_LEN -BEAM_TUBE_WIDTH/2 - CORNER_TUBE_SHIFT,
			   CORNER_TUBE_SHIFT + TUBE_WIDTH + TRANSVERSE_TUBE_SHIFT,
			   CORNER_HEIGHT - BEAM_TUBE_WIDTH*.55])
	rotate([270, 0, 0])
	cylinder(h=CORNER_LEN+1, r=BEAM_TUBE_WIDTH/2);
}

module bl_corner_beam() {
	difference() {
		bl_corner_piece_smooth3d();
		frame_tube_x();
		transverse_tube();
		frame_tube_y();
	}
}

////////////
// Double //

// Manual edition of the STL is necessary to remove the protrusions where the mirrored corners meet,
// they could also be removed here but the ROI is too low.
module bl_double_corner_beam() {
	translate([-2* FRAME_TUBE_OFFSET, 0, 0])
	bl_corner_beam();
	mirror([1, 0, 0])
		bl_corner_beam();
}

//////////
// Foot //
// TPU foot to put below each cornerstone for added grip and to avoid scraping the floor.

module foot() {
	difference() {
		cylinder(h=SIZE/3 + FOOT_OFFSET, r=SIZE/2);
		translate([0, 0, SIZE/2 + FOOT_OFFSET])
		truncube(SIZE * FOOT_TRUNCUBE_VOID_SCALE, TRUNCATION_RATIO);
	}
}

////////////////////
// Shape assembly //
////////////////////

// cornerstone();
// support();

translate([0, -CORNER_LEN - 50,  0])
bl_corner_beam();
/*
translate([200, 0, 0])
bl_double_corner_beam();

translate([0, 100, 0])
foot();
*/