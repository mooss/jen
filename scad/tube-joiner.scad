// Size of the cube forming the core of the tube joiner.
SIZE = 30;

// How deep the tube goes inside a given face.
TUBE_DEPTH = 8;

// Witdh of the tube.
TUBE_WIDTH = 9.75;

// How much should the base cube be truncated.
TRUNCATION_RATIO = .75;

// Height of the support circle.
SUPPORT_HEIGHT = .4;

// Radius of the support circle.
SUPPORT_RADIUS = 15;

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
module joiner() {
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

joiner();
support();
