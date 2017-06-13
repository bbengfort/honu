#!/bin/bash

for i in {1..32}
   do
       # Lin, Single Key
       fab experiment:relax=False,multikey=False,procs=$i

       # Lin, Multi Key
       fab experiment:relax=False,multikey=True,procs=$i

       # Seq, Single Key
       fab experiment:relax=True,multikey=False,procs=$i

       # Seq, Multi Key
       fab experiment:relax=True,multikey=True,procs=$i
done
