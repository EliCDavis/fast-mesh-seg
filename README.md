# Fast Mesh Segmentation

Meant to take large files (multiple GBs worth / 100million+ tris) and segment them into multiple smaller files for sake of loading them into game engines.

## Features

Why this is potentially better than using other programs:

* Only loads what FBX nodes are needed for mesh segmentation. Ignores all other fbx data, saving on RAM and loading time. 
* Delays uncompressing array-type properties until needed, uncompression occurs in worker pool.
* Geometry Nodes are streamed to a worker pool as they are pulled from the file. Splitting the mesh is multithreaded and begins before the FBX file is done being read.

## Example Output

![Results](https://i.imgur.com/QCW2qzq.png)

## Progress History / Lessons Learned

### Niave Implementation

Just trying to get something working

```txt
2020/01/02 22:43:05 Loading Model: dragon_vrip.fbx took 517.0337ms
2020/01/02 22:43:05 Splitting model by plane took 454.0025ms
2020/01/02 22:43:08 Saving Model (287745 tris) as 'retained.obj' took 2.3630315s
2020/01/02 22:43:13 Saving Model (578630 tris) as 'clipped.obj' took 5.0589994s
```

```txt
2020/01/03 22:22:20 Loading Model: HIB-model.fbx took 2m29.4960692s
2020/01/03 22:22:39 Splitting model by plane took 18.371965s
2020/01/03 22:22:39 Retained Model Polygon Count: 28562401
2020/01/03 22:22:39 Clipped Model Polygon Count: 9922739
```

### Defering Loading Properties until their needed:

You see that loading the FBX is almost instant now for the dragon, but it takes longer to split the model because we have to now decompress the nodes we need, steps that where originally taken care of during the FBX loading step. 

```txt
2020/01/04 01:39:34 Loading Model: dragon_vrip.fbx took 8.0032ms
2020/01/04 01:39:35 Splitting model by plane took 799.9983ms
2020/01/04 01:39:37 Saving Model (287745 tris) as 'retained.obj' took 2.4430214s
2020/01/04 01:39:42 Saving Model (578630 tris) as 'clipped.obj' took 4.5939848s
```

```txt
2020/01/04 01:45:19 Loading Model: HIB-model.fbx took 2m9.1930661s
2020/01/04 01:45:52 Splitting model by plane took 32.672s
2020/01/04 01:45:52 Retained Model Polygon Count: 28562401
2020/01/04 01:45:52 Clipped Model Polygon Count: 9922739
```

### Efficiently Interpreting Number Types / Minimizing Array Resizing / SeekReader

Previously, the FBX reader used `binary.Read` method in golang. Doing so required making small readers and having the method use a switch statement to try to determine what number type it was dealing with. Creating these small readers where wasteful since we already have the data loaded, and we already know what the number type is so we don't have to go through a wasteful switch statement. This resulted with the most speadup for the loading of our large model (`HIB-model.fbx`) and spead it up a few seconds.

Minimizing array resizing when splitting the geometry nodes involved creating and re-using larger sized arrays instead of just appending to one each face. This means that we do minimal array resizing but we have to guess the size of the array beforehand because we don't know how many polygons will exist on each side of the clipping plane until we've completed the cutting operation. This resulted in the most amount of speedup (1.5x) for our small model that only has 1 geometry node. This ended up slowing down our model splitting for our big model (the opposite of what we are going for) by 1 second. This is because it has a  large amount of geometry nodes, which means theres a lot of wasted array space I guess. Further investigation and research is required.

```txt
-> Loading and splitting model by plane took 533.0014ms
----> Loading Model: dragon_vrip.fbx took 6.036ms
----> Splitting model by plane took 526.9654ms
2020/01/04 17:40:50 Retained Model Polygon Count: 287745
2020/01/04 17:40:50 Clipped Model Polygon Count: 578630
```

```txt
-> Loading and splitting model by plane took 2m38.8481157s
----> Loading Model: HIB-model.fbx took 2m5.1550347s
----> Splitting model by plane took 33.5160793s
2020/01/04 17:43:28 Retained Model Polygon Count: 28562401
2020/01/04 17:43:28 Clipped Model Polygon Count: 9922739
```

### Streaming Geometry Nodes to Worker Pool

Instead of waiting for the entire FBX file to be read in, we send geomatry nodes immediately after they've been read to a worker pool that pass some matcher function. This means we can start splitting the geometry before we've even finished reading the file, and that splitting is done over multiple threads (the number dependent on the machine the program is being ran on). Doing this spead up both our small file and large file benchmarks. It was a little disapoiting how little the speedup we recieved for the large file (1.08x), and that's probably due to how many small geometry nodes exist within it.

One reason the large file isn't getting that large of a speedup is because with the introduction of channels, there comes an associated communication cost. Because there are a lot of very small geometry nodes (304905 of them), there's a lot of communication overhead for splitting up very easy tasks. I imagine you would experience a much larger speedup with larger geomeetry nodes type files. This issue can hopefully be remedied by batching nodes as a single job instead of sending them one at a time.

```txt
-> Loading and splitting dragon_vrip.fbx by plane took 523.9953ms
2020/01/05 00:41:36 Retained Model Polygon Count: 287745
2020/01/05 00:41:36 Clipped Model Polygon Count: 578630
```

```txt
-> Loading and splitting HIB-model.fbx by plane took 2m26.6470021s
2020/01/05 00:39:39 Retained Model Polygon Count: 28562401
2020/01/05 00:39:39 Clipped Model Polygon Count: 9922739
```

### Chunking Jobs + More Efficient Node Header Loading

Now that we have multiple workers, we want them to spend as much as their time working instead of communicating. That means we need to give them large jobs that keep them occupied for long times, but also at the same time not keeping them waiting too long for their next job. I have the rearder build a list of nodes and keep up with how large each node is that's read in. As soon as the total size of all collected nodes reach a certain threshold, the list of nodes is sent to a single worker. I found a good threshold to be
1000000 bytes (~1MB) and to have 3 Workers (my machine only has 4 cores).  These parameters of course are super specific to my machine and the test file being ran. I need to develop a program for creating a variaty of FBX with different properties for proper benchmarking. Overall this dropped the time it took for splitting the larger mesh by 10 whole seconds. The smaller single geometry node file now take a little longer to complete which is expected. During the read process it never breaks the threshold to be sent to a job, and thus is split once the entire FBX file is done being loaded in. Not sure how to remedy this, maybe some how provide the reader with the file size before it begins so it can adjust the size theshold for sening nodes to jobs?

The MAJOR speedup came from benchmarking and viewing the cpu profile output.

![Profiling Output](https://i.imgur.com/yXqwhR6.png)

You can see a majority of the time spent parsing a node is on it's header data (how big it is, how many properties it has, etc). These are just a few simple int32/int64 values that always come in the same exact order and come in every single node. So instead of performing multiple reads per node for it's header, I just perform one single read and construct the values directly from the bytes read in. This resulted in a MASSIVE speed up of 50 seconds (1.6x). I am very happy. I need to explore more places where arrays can easily be reused for reading.

```txt
-> Loading and splitting dragon_vrip.fbx by plane with 3 workers took 662.1468ms
2020/01/05 16:42:19 Retained Model Polygon Count: 287745
2020/01/05 16:42:19 Clipped Model Polygon Count: 578630
```

```txt
-> Loading and splitting HIB-model.fbx by plane with 3 workers took 1m25.5735363s
2020/01/05 16:39:00 Retained Model Polygon Count: 28562401
2020/01/05 16:39:00 Clipped Model Polygon Count: 9891243
```

## Roadmap

* [x] Outputting basic splitting of fbx file into multiple models
* [ ] Recursively Build Octree based on desired polycount threshold
* [x] Stream polygons as their unpackaged from geometry instead of reading entire fbx file first.
* [ ] Feed Poly stream into CUDA
* [ ] Create program for generating different size FBX with different number of geometry nodes and node sizes.

## Credits

### OSS 

Credits to [o5h](https://github.com/o5h/fbx/tree/3a77542940a3e1fb404bfd00f2e49565a504a2df) and [three.js](https://github.com/mrdoob/three.js/blob/de530d6bae1bf40d1e001411bc3e02a915c2c993/examples/js/loaders/FBXLoader.js) for helping me get started with code. Credits to blender for a detailed [explanation to the fbx format](https://code.blender.org/2013/08/fbx-binary-file-format-specification/). Also [this repository](https://github.com/hamish-milne/FbxWriter/blob/master/Fbx/FbxBinary.cs) for helping me with the writer.

### Other Resources

* [Clipping a Mesh Against a Plane](https://www.geometrictools.com/Documentation/ClipMesh.pdf)
* [Binary Diffing](https://superuser.com/questions/125376/how-do-i-compare-binary-files-in-linux/968863#968863) (for debugging purposes).