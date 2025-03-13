module octahedron(size) {
	points=[[ 1,  0,  0],  // Right.
			[ 0,  1,  0],  // Forward.
			[-1,  0,  0],  // Left.
			[ 0, -1,  0],  // Backward.
			[ 0,  0,  1],  // Up.
			[ 0,  0, -1]]; // Down.

	faces=[[4, 1, 0], // Top faces.
		   [4, 2, 1],
		   [4, 3, 2],
		   [4, 0, 3],
		   [5, 0, 1], // Bottom faces.
		   [5, 1, 2],
		   [5, 2, 3],
		   [5, 3, 0]];

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

truncube(10, .75);
